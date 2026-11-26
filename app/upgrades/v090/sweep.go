package v090

import (
	"fmt"
	"slices"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"
)

const (
	EventTypeRecoverySweep = "recovery_sweep"
	AttributeKeyAddress    = "address"
)

// Sweep moves to cfg's recipient every coin compromised addresses hold, after turning vesting accounts
// into base accounts and releasing delegations and unbonding delegations. It fails without changing
// state when the recipient or a listed address is a module account or a contract.
func Sweep(ctx sdk.Context, k *keepers.AppKeepers, cfg ChainConfig) error {
	if err := checkSweepable(ctx, k.AccountKeeper, k.BankKeeper, &k.WasmKeeper, cfg); err != nil {
		return err
	}
	bondDenom, err := k.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return fmt.Errorf("read bond denom: %w", err)
	}

	total := sdk.NewCoins()
	for _, addr := range cfg.compromised {
		moved, err := sweepCompromised(ctx, k.AccountKeeper, k.DistrKeeper, k.StakingKeeper, k.BankKeeper, bondDenom, addr, cfg.recipient)
		if err != nil {
			return fmt.Errorf("sweep compromised %s: %w", addr, err)
		}
		emitSwept(ctx, addr, moved)
		total = total.Add(moved...)
	}

	ctx.Logger().Info("recovery sweep done",
		"recipient", cfg.recipient.String(),
		"addresses", len(cfg.compromised),
		"swept", total.String(),
	)
	return nil
}

func checkSweepable(
	ctx sdk.Context, accounts *authkeeper.AccountKeeper, bank *bankkeeper.BaseKeeper, wasm *wasmkeeper.Keeper, cfg ChainConfig,
) error {
	for _, addr := range slices.Concat([]sdk.AccAddress{cfg.recipient}, cfg.compromised) {
		if bank.BlockedAddr(addr) {
			return fmt.Errorf("%w: %s is a blocked address", ErrProtectedAddress, addr)
		}
		if _, isModule := accounts.GetAccount(ctx, addr).(sdk.ModuleAccountI); isModule {
			return fmt.Errorf("%w: %s is a module account", ErrProtectedAddress, addr)
		}
		if wasm.HasContractInfo(ctx, addr) {
			return fmt.Errorf("%w: %s is a contract", ErrProtectedAddress, addr)
		}
	}
	return nil
}

func sweepCompromised(
	ctx sdk.Context, accounts *authkeeper.AccountKeeper, distr *distrkeeper.Keeper, staking *stakingkeeper.Keeper,
	bank *bankkeeper.BaseKeeper, bondDenom string, addr, recipient sdk.AccAddress,
) (sdk.Coins, error) {
	unlockVesting(ctx, accounts, addr)
	// rewards withdrawn by the unbond hooks must reach addr, not an attacker-set withdraw address
	if err := distr.DeleteDelegatorWithdrawAddr(ctx, addr, addr); err != nil {
		return nil, fmt.Errorf("reset withdraw address: %w", err)
	}

	var delegations []stakingtypes.Delegation
	if err := staking.IterateDelegatorDelegations(ctx, addr, func(d stakingtypes.Delegation) bool {
		delegations = append(delegations, d)
		return false
	}); err != nil {
		return nil, fmt.Errorf("read delegations: %w", err)
	}
	for _, d := range delegations {
		if err := unbondDelegation(ctx, staking, bank, bondDenom, addr, d); err != nil {
			return nil, fmt.Errorf("unbond from %s: %w", d.ValidatorAddress, err)
		}
	}

	var ubds []stakingtypes.UnbondingDelegation
	if err := staking.IterateDelegatorUnbondingDelegations(ctx, addr, func(ubd stakingtypes.UnbondingDelegation) bool {
		ubds = append(ubds, ubd)
		return false
	}); err != nil {
		return nil, fmt.Errorf("read unbonding delegations: %w", err)
	}
	for _, ubd := range ubds {
		if err := completeUnbonding(ctx, staking, addr, ubd); err != nil {
			return nil, fmt.Errorf("complete unbonding from %s: %w", ubd.ValidatorAddress, err)
		}
	}

	spendable := bank.SpendableCoins(ctx, addr)
	if spendable.IsZero() {
		return spendable, nil
	}
	if err := bank.SendCoins(ctx, addr, recipient, spendable); err != nil {
		return nil, fmt.Errorf("send %s to recipient: %w", spendable, err)
	}
	return spendable, nil
}

func unlockVesting(ctx sdk.Context, accounts *authkeeper.AccountKeeper, addr sdk.AccAddress) {
	vacc, ok := accounts.GetAccount(ctx, addr).(vestingexported.VestingAccount)
	if !ok {
		return
	}
	ctx.Logger().Info("recovery sweep unlocks vesting account",
		"address", addr.String(),
		"vesting", vacc.GetVestingCoins(ctx.BlockTime()).String(),
		"original_vesting", vacc.GetOriginalVesting().String(),
		"start_time", vacc.GetStartTime(),
		"end_time", vacc.GetEndTime(),
	)
	accounts.SetAccount(ctx, authtypes.NewBaseAccount(addr, vacc.GetPubKey(), vacc.GetAccountNumber(), vacc.GetSequence()))
}

func unbondDelegation(
	ctx sdk.Context, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper,
	bondDenom string, addr sdk.AccAddress, d stakingtypes.Delegation,
) error {
	valAddr, err := staking.ValidatorAddressCodec().StringToBytes(d.ValidatorAddress)
	if err != nil {
		return fmt.Errorf("decode validator address: %w", err)
	}
	// read before Unbond, which removes an unbonded validator left without shares
	validator, err := staking.GetValidator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("read validator: %w", err)
	}
	pool := stakingtypes.NotBondedPoolName
	if validator.IsBonded() {
		pool = stakingtypes.BondedPoolName
	}

	amount, err := staking.Unbond(ctx, addr, valAddr, d.Shares)
	if err != nil {
		return fmt.Errorf("unbond shares: %w", err)
	}
	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, amount))
	// Undelegate without the unbonding entry: pay straight from the pool the tokens sit in; dust shares
	// worth zero tokens leave coins empty, which the bank moves as a no-op
	if err := bank.UndelegateCoinsFromModuleToAccount(ctx, pool, addr, coins); err != nil {
		return fmt.Errorf("pay %s from %s: %w", coins, pool, err)
	}
	return nil
}

func completeUnbonding(ctx sdk.Context, staking *stakingkeeper.Keeper, addr sdk.AccAddress, ubd stakingtypes.UnbondingDelegation) error {
	valAddr, err := staking.ValidatorAddressCodec().StringToBytes(ubd.ValidatorAddress)
	if err != nil {
		return fmt.Errorf("decode validator address: %w", err)
	}
	for i := range ubd.Entries {
		// the queue item left at the original time later finds no unbonding delegation and is skipped
		ubd.Entries[i].CompletionTime = ctx.BlockTime()
	}
	if err := staking.SetUnbondingDelegation(ctx, ubd); err != nil {
		return fmt.Errorf("mature entries: %w", err)
	}
	if _, err := staking.CompleteUnbonding(ctx, addr, valAddr); err != nil {
		return fmt.Errorf("complete: %w", err)
	}
	return nil
}

func emitSwept(ctx sdk.Context, addr sdk.AccAddress, moved sdk.Coins) {
	if moved.IsZero() {
		return
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(EventTypeRecoverySweep,
		sdk.NewAttribute(AttributeKeyAddress, addr.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, moved.String()),
	))
}
