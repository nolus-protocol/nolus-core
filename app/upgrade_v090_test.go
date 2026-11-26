package app

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	v090 "github.com/Nolus-Protocol/nolus-core/app/upgrades/v090"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

const (
	sweepBondDenom = params.DefaultBondDenom
	sweepOtherDnm  = "uosmo"
)

type sweepFixture struct {
	app       *App
	ctx       sdk.Context
	staking   stakingtypes.MsgServer
	validator sdk.ValAddress
}

func sweepAddr(b byte) sdk.AccAddress {
	return bytes.Repeat([]byte{b}, 20)
}

func newSweepFixture(t *testing.T) sweepFixture {
	t.Helper()
	params.GetDefaultConfig()
	testApp, ctx := CreateTestApp(true, t.TempDir())
	ctx = ctx.WithChainID("sweep-1").
		WithBlockHeight(10).
		WithBlockTime(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)).
		WithEventManager(sdk.NewEventManager())

	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = sweepBondDenom
	require.NoError(t, testApp.StakingKeeper.SetParams(ctx, stakingParams), "staking params are stored")
	require.NoError(t, testApp.DistrKeeper.Params.Set(ctx, distrtypes.DefaultParams()), "distribution params are stored")
	require.NoError(t, testApp.DistrKeeper.FeePool.Set(ctx, distrtypes.InitialFeePool()), "the fee pool is stored")
	require.NoError(t, testApp.SlashingKeeper.SetParams(ctx, slashingtypes.DefaultParams()), "slashing params are stored")
	require.NoError(t, testApp.WasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()), "wasm params are stored")

	f := sweepFixture{app: testApp, ctx: ctx, staking: stakingkeeper.NewMsgServerImpl(testApp.StakingKeeper)}

	operator := sweepAddr(0xEE)
	f.fund(t, operator, sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 10_000_000)))
	f.validator = f.createValidator(t, operator, 10_000_000)
	return f
}

func (f sweepFixture) createValidator(t *testing.T, operator sdk.AccAddress, selfBond int64) sdk.ValAddress {
	t.Helper()
	valAddr := sdk.ValAddress(operator)
	msg, err := stakingtypes.NewMsgCreateValidator(
		f.valString(t, valAddr),
		ed25519.GenPrivKey().PubKey(),
		sdk.NewInt64Coin(sweepBondDenom, selfBond),
		stakingtypes.NewDescription("validator", "", "", "", ""),
		stakingtypes.NewCommissionRates(
			sdkmath.LegacyNewDecWithPrec(1, 1), sdkmath.LegacyNewDecWithPrec(2, 1), sdkmath.LegacyNewDecWithPrec(1, 2),
		),
		sdkmath.OneInt(),
	)
	require.NoError(t, err, "the create-validator message is well formed")
	_, err = f.staking.CreateValidator(f.ctx, msg)
	require.NoError(t, err, "the validator is created")
	return valAddr
}

func (f sweepFixture) delegate(t *testing.T, del sdk.AccAddress, val sdk.ValAddress, amount int64) {
	t.Helper()
	_, err := f.staking.Delegate(f.ctx, stakingtypes.NewMsgDelegate(f.accString(t, del), f.valString(t, val), sdk.NewInt64Coin(sweepBondDenom, amount)))
	require.NoError(t, err, "the delegation succeeds")
}

func (f sweepFixture) validatorState(t *testing.T, val sdk.ValAddress) stakingtypes.Validator {
	t.Helper()
	v, err := f.app.StakingKeeper.GetValidator(f.ctx, val)
	require.NoError(t, err, "the validator exists")
	return v
}

func (f sweepFixture) fund(t *testing.T, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()
	require.NoError(t, f.app.BankKeeper.MintCoins(f.ctx, minttypes.ModuleName, coins), "test coins are minted")
	require.NoError(t, f.app.BankKeeper.SendCoinsFromModuleToAccount(f.ctx, minttypes.ModuleName, addr, coins), "the test account is funded")
}

func (f sweepFixture) accString(t *testing.T, addr sdk.AccAddress) string {
	t.Helper()
	s, err := f.app.AccountKeeper.AddressCodec().BytesToString(addr)
	require.NoError(t, err, "the account address encodes to bech32")
	return s
}

func (f sweepFixture) valString(t *testing.T, val sdk.ValAddress) string {
	t.Helper()
	s, err := f.app.StakingKeeper.ValidatorAddressCodec().BytesToString(val)
	require.NoError(t, err, "the validator address encodes to bech32")
	return s
}

func (f sweepFixture) endBlock(t *testing.T) {
	t.Helper()
	_, err := f.app.StakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err, "the staking end blocker succeeds")
}

func (f sweepFixture) balances(addrs ...sdk.AccAddress) []string {
	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		out = append(out, f.app.BankKeeper.GetAllBalances(f.ctx, addr).String())
	}
	return out
}

func (f sweepFixture) pools() []sdk.AccAddress {
	return []sdk.AccAddress{
		f.app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName),
		f.app.AccountKeeper.GetModuleAddress(stakingtypes.NotBondedPoolName),
	}
}

func (f sweepFixture) sweepEvents() []sdk.Event {
	var out []sdk.Event
	for _, ev := range f.ctx.EventManager().Events() {
		if ev.Type == v090.EventTypeRecoverySweep {
			out = append(out, ev)
		}
	}
	return out
}

func (f sweepFixture) sweepEvent(t *testing.T, addr sdk.AccAddress, amount sdk.Coins) sdk.Event {
	t.Helper()
	return sdk.NewEvent(v090.EventTypeRecoverySweep,
		sdk.NewAttribute(v090.AttributeKeyAddress, f.accString(t, addr)),
		sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
	)
}

func (f sweepFixture) requirePoolsBacked(t *testing.T) {
	t.Helper()
	bonded, notBonded := sdkmath.ZeroInt(), sdkmath.ZeroInt()
	validators, err := f.app.StakingKeeper.GetAllValidators(f.ctx)
	require.NoError(t, err, "the validators are readable")
	for _, v := range validators {
		if v.IsBonded() {
			bonded = bonded.Add(v.Tokens)
		} else {
			notBonded = notBonded.Add(v.Tokens)
		}
	}
	require.NoError(t, f.app.StakingKeeper.IterateUnbondingDelegations(f.ctx, func(_ int64, ubd stakingtypes.UnbondingDelegation) bool {
		for _, e := range ubd.Entries {
			notBonded = notBonded.Add(e.Balance)
		}
		return false
	}), "the unbonding delegations are readable")

	pools := f.pools()
	require.Equal(t, bonded.String(), f.app.BankKeeper.GetBalance(f.ctx, pools[0], sweepBondDenom).Amount.String(),
		"bonded pool must hold the bonded validators' tokens")
	require.Equal(t, notBonded.String(), f.app.BankKeeper.GetBalance(f.ctx, pools[1], sweepBondDenom).Amount.String(),
		"not-bonded pool must hold the unbonding entries and the not-bonded validators' tokens")
}

func mustSweepConfig(t *testing.T, recipient sdk.AccAddress, compromised []sdk.AccAddress) v090.ChainConfig {
	t.Helper()
	cfg, err := v090.NewChainConfig(recipient, compromised)
	require.NoError(t, err, "the test sweep config is valid")
	return cfg
}

func TestV090IsRegisteredWithoutStoreChanges(t *testing.T) {
	require.Equal(t, "v0.9.0", v090.UpgradeName, "the on-chain plan name must be v0.9.0")
	require.Equal(t, storetypes.StoreUpgrades{Added: []string{}, Deleted: []string{}}, v090.Upgrade.StoreUpgrades,
		"v0.9.0 adds, deletes and renames no stores")

	names := make([]string, 0, len(Upgrades))
	for _, u := range Upgrades {
		names = append(names, u.UpgradeName)
	}
	require.Contains(t, names, v090.UpgradeName, "v0.9.0 must be registered in the Upgrades list")

	testApp, _ := CreateTestApp(true, t.TempDir())
	require.True(t, testApp.UpgradeKeeper.HasHandler(v090.UpgradeName),
		"the upgrade keeper must hold a handler for v0.9.0")
}

func TestV090SweepReleasesCompromisedStake(t *testing.T) {
	f := newSweepFixture(t)
	compromised, attacker, recipient := sweepAddr(1), sweepAddr(2), sweepAddr(3)
	liquid := sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 1_000_000), sdk.NewInt64Coin(sweepOtherDnm, 777))
	f.fund(t, compromised, liquid.Add(sdk.NewInt64Coin(sweepBondDenom, 5_000_000)))

	f.delegate(t, compromised, f.validator, 5_000_000)
	f.endBlock(t)
	validator := f.validatorState(t, f.validator)
	require.True(t, validator.IsBonded(), "the validator must be bonded before the sweep")

	undelegated, err := f.staking.Undelegate(f.ctx, stakingtypes.NewMsgUndelegate(
		f.accString(t, compromised), f.valString(t, f.validator), sdk.NewInt64Coin(sweepBondDenom, 2_000_000)))
	require.NoError(t, err, "the undelegation succeeds")
	require.NoError(t, f.app.DistrKeeper.SetDelegatorWithdrawAddr(f.ctx, compromised, attacker), "the attacker withdraw address is stored")

	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)
	rewardPool := sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 1_300_000))
	require.NoError(t, f.app.BankKeeper.MintCoins(f.ctx, minttypes.ModuleName, rewardPool), "reward coins are minted")
	require.NoError(t, f.app.BankKeeper.SendCoinsFromModuleToModule(f.ctx, minttypes.ModuleName, distrtypes.ModuleName, rewardPool), "the distribution module is funded")
	require.NoError(t, f.app.DistrKeeper.AllocateTokensToValidator(f.ctx, validator, sdk.NewDecCoinsFromCoins(rewardPool...)), "rewards are allocated to the validator")

	probe, _ := f.ctx.CacheContext()
	rewards, err := f.app.DistrKeeper.WithdrawDelegationRewards(probe, compromised, f.validator)
	require.NoError(t, err, "the pending rewards are withdrawable")
	require.False(t, rewards.IsZero(), "the compromised delegation must have pending rewards")

	cfg := mustSweepConfig(t, recipient, []sdk.AccAddress{compromised})
	require.NoError(t, v090.Sweep(f.ctx, &f.app.AppKeepers, cfg), "the sweep succeeds")

	delegations, err := f.app.StakingKeeper.GetDelegatorDelegations(f.ctx, compromised, 10)
	require.NoError(t, err, "the delegations are readable")
	require.Empty(t, delegations, "the compromised address keeps no delegation")
	ubds, err := f.app.StakingKeeper.GetUnbondingDelegations(f.ctx, compromised, 10)
	require.NoError(t, err, "the unbonding delegations are readable")
	require.Empty(t, ubds, "the compromised address keeps no unbonding delegation")

	swept := liquid.Add(sdk.NewInt64Coin(sweepBondDenom, 3_000_000), undelegated.Amount).Add(rewards...)
	require.Equal(t, []string{"", "", swept.String()}, f.balances(compromised, attacker, recipient),
		"the recipient receives every liquid, delegated, unbonding and reward coin, and the attacker nothing")
	require.Equal(t, []sdk.Event{f.sweepEvent(t, compromised, swept)}, f.sweepEvents(),
		"the sweep emits one event for the compromised address")
	f.requirePoolsBacked(t)

	tracked := []sdk.AccAddress{compromised, attacker, recipient, f.pools()[0], f.pools()[1]}
	before := f.balances(tracked...)
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1).WithBlockTime(undelegated.CompletionTime.Add(time.Hour))
	f.endBlock(t)
	require.Equal(t, before, f.balances(tracked...), "the stale unbonding queue item moves nothing when it falls due")
	f.requirePoolsBacked(t)
}

func TestV090SweepUnlocksVestingAccounts(t *testing.T) {
	f := newSweepFixture(t)
	f.endBlock(t)
	pubKey := secp256k1.GenPrivKey().PubKey()
	compromised, recipient := sdk.AccAddress(pubKey.Address()), sweepAddr(3)
	f.fund(t, compromised, sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 2_000_000)))
	base, ok := f.app.AccountKeeper.GetAccount(f.ctx, compromised).(*authtypes.BaseAccount)
	require.True(t, ok, "the funded address is a base account")
	require.NoError(t, base.SetPubKey(pubKey), "the account takes a public key")
	require.NoError(t, base.SetSequence(7), "the account takes a sequence")
	now := f.ctx.BlockTime()
	vesting, err := vestingtypes.NewContinuousVestingAccount(base, sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 2_000_000)),
		now.AddDate(-1, 0, 0).Unix(), now.AddDate(1, 0, 0).Unix())
	require.NoError(t, err, "the continuous vesting account is valid")
	f.app.AccountKeeper.SetAccount(f.ctx, vesting)
	require.False(t, f.app.BankKeeper.LockedCoins(f.ctx, compromised).IsZero(), "the account is mid-schedule with locked coins")

	f.delegate(t, compromised, f.validator, 1_000_000)
	_, err = f.staking.Undelegate(f.ctx, stakingtypes.NewMsgUndelegate(
		f.accString(t, compromised), f.valString(t, f.validator), sdk.NewInt64Coin(sweepBondDenom, 400_000)))
	require.NoError(t, err, "the undelegation succeeds")

	cfg := mustSweepConfig(t, recipient, []sdk.AccAddress{compromised})
	require.NoError(t, v090.Sweep(f.ctx, &f.app.AppKeepers, cfg), "the sweep succeeds")

	require.Equal(t, authtypes.NewBaseAccount(compromised, pubKey, base.GetAccountNumber(), 7),
		f.app.AccountKeeper.GetAccount(f.ctx, compromised),
		"the vesting account becomes a base account with the same public key, number and sequence")
	require.Equal(t, []string{"", sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 2_000_000)).String()},
		f.balances(compromised, recipient), "the recipient receives every coin, the formerly locked ones included")
	f.requirePoolsBacked(t)
}

func TestV090SweepUnbondsFromEveryValidatorStatus(t *testing.T) {
	compromised, recipient := sweepAddr(1), sweepAddr(3)
	const staked = 4_000_000

	tests := []struct {
		name  string
		setup func(t *testing.T, f sweepFixture) sdk.ValAddress
		// wantValidator derives the validator expected after the sweep from its state before; nil means removed
		wantValidator func(before stakingtypes.Validator) *stakingtypes.Validator
	}{
		{
			name: "unbonding jailed validator",
			setup: func(t *testing.T, f sweepFixture) sdk.ValAddress {
				t.Helper()
				f.fund(t, compromised, sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, staked)))
				operator := sweepAddr(0xE1)
				f.fund(t, operator, sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 10_000_000)))
				val := f.createValidator(t, operator, 10_000_000)
				f.endBlock(t)
				f.delegate(t, compromised, val, staked)
				consAddr, err := f.validatorState(t, val).GetConsAddr()
				require.NoError(t, err, "the validator has a consensus address")
				require.NoError(t, f.app.StakingKeeper.Jail(f.ctx, consAddr), "the validator is jailed")
				f.endBlock(t)
				require.True(t, f.validatorState(t, val).IsUnbonding(), "the jailed validator is unbonding before the sweep")
				return val
			},
			wantValidator: func(before stakingtypes.Validator) *stakingtypes.Validator {
				before.Tokens = sdkmath.NewInt(10_000_000)
				before.DelegatorShares = sdkmath.LegacyNewDec(10_000_000)
				return &before
			},
		},
		{
			name: "unbonded validator whose only delegator is the compromised address",
			setup: func(t *testing.T, f sweepFixture) sdk.ValAddress {
				t.Helper()
				f.fund(t, compromised, sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, staked)))
				operator := sweepAddr(0xE2)
				f.fund(t, operator, sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 2_000_000)))
				val := f.createValidator(t, operator, 2_000_000)
				f.delegate(t, compromised, val, staked)
				_, err := f.staking.Undelegate(f.ctx, stakingtypes.NewMsgUndelegate(
					f.accString(t, operator), f.valString(t, val), sdk.NewInt64Coin(sweepBondDenom, 2_000_000)))
				require.NoError(t, err, "the operator undelegates its whole self-delegation")
				require.True(t, f.validatorState(t, val).IsUnbonded(), "the validator is unbonded before the sweep")
				return val
			},
			wantValidator: func(stakingtypes.Validator) *stakingtypes.Validator { return nil },
		},
		{
			name: "bonded validator operated by the compromised address",
			setup: func(t *testing.T, f sweepFixture) sdk.ValAddress {
				t.Helper()
				f.fund(t, compromised, sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, staked)))
				val := f.createValidator(t, compromised, staked)
				f.endBlock(t)
				require.True(t, f.validatorState(t, val).IsBonded(), "the compromised validator is bonded before the sweep")
				return val
			},
			wantValidator: func(before stakingtypes.Validator) *stakingtypes.Validator {
				before.Jailed = true
				before.Tokens = sdkmath.ZeroInt()
				before.DelegatorShares = sdkmath.LegacyZeroDec()
				return &before
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newSweepFixture(t)
			f.endBlock(t)
			val := tc.setup(t, f)
			before := f.validatorState(t, val)
			f.requirePoolsBacked(t)

			cfg := mustSweepConfig(t, recipient, []sdk.AccAddress{compromised})
			require.NoError(t, v090.Sweep(f.ctx, &f.app.AppKeepers, cfg), "the sweep succeeds")

			swept := sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, staked)).String()
			require.Equal(t, []string{"", swept}, f.balances(compromised, recipient), "the recipient receives every staked coin")
			delegations, err := f.app.StakingKeeper.GetDelegatorDelegations(f.ctx, compromised, 10)
			require.NoError(t, err, "the delegations are readable")
			require.Empty(t, delegations, "the compromised address keeps no delegation")
			got, err := f.app.StakingKeeper.GetValidator(f.ctx, val)
			if want := tc.wantValidator(before); want == nil {
				require.ErrorIs(t, err, stakingtypes.ErrNoValidatorFound, "the unbond removes the validator left without shares")
			} else {
				require.NoError(t, err, "the validator still exists")
				require.Equal(t, *want, got, "the validator loses exactly the compromised stake")
			}
			f.requirePoolsBacked(t)

			unbondingTime, err := f.app.StakingKeeper.UnbondingTime(f.ctx)
			require.NoError(t, err, "the unbonding time is readable")
			f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1).WithBlockTime(f.ctx.BlockTime().Add(unbondingTime + time.Hour))
			f.endBlock(t)
			require.Equal(t, []string{"", swept}, f.balances(compromised, recipient), "later blocks move nothing more")
			f.requirePoolsBacked(t)
		})
	}
}

func TestV090SweepMovesNothingFromEmptyAddresses(t *testing.T) {
	f := newSweepFixture(t)
	empty, recipient := []sdk.AccAddress{sweepAddr(1), sweepAddr(4)}, sweepAddr(3)

	cfg := mustSweepConfig(t, recipient, empty)
	require.NoError(t, v090.Sweep(f.ctx, &f.app.AppKeepers, cfg), "the sweep succeeds")
	require.Equal(t, []string{"", "", ""}, f.balances(empty[0], empty[1], recipient), "no coins move")
	require.Empty(t, f.sweepEvents(), "the sweep emits no event for addresses that moved nothing")
}

func TestV090SweepRefusesProtectedAddresses(t *testing.T) {
	victim, recipient := sweepAddr(1), sweepAddr(3)
	victimFunds := sdk.NewCoins(sdk.NewInt64Coin(sweepBondDenom, 1_000))

	tests := []struct {
		name      string
		protected func(t *testing.T, f sweepFixture) sdk.AccAddress
		recipient bool
	}{
		{
			name: "blocked module address listed",
			protected: func(*testing.T, sweepFixture) sdk.AccAddress {
				return authtypes.NewModuleAddress(distrtypes.ModuleName)
			},
		},
		{
			name: "unblocked module account listed",
			protected: func(t *testing.T, f sweepFixture) sdk.AccAddress {
				t.Helper()
				macc := authtypes.NewEmptyModuleAccount("recovery-test")
				f.app.AccountKeeper.SetAccount(f.ctx, f.app.AccountKeeper.NewAccount(f.ctx, macc))
				require.False(t, f.app.BankKeeper.BlockedAddr(macc.GetAddress()),
					"the module account must not be blocked so the account-type check is the one that refuses it")
				return macc.GetAddress()
			},
		},
		{
			name:      "contract listed",
			protected: contractAt20Bytes,
		},
		{
			name:      "contract as recipient",
			protected: contractAt20Bytes,
			recipient: true,
		},
		{
			name: "module account as recipient",
			protected: func(*testing.T, sweepFixture) sdk.AccAddress {
				return authtypes.NewModuleAddress(distrtypes.ModuleName)
			},
			recipient: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newSweepFixture(t)
			f.fund(t, victim, victimFunds)
			protected := tc.protected(t, f)

			cfg := mustSweepConfig(t, recipient, []sdk.AccAddress{victim, protected})
			if tc.recipient {
				cfg = mustSweepConfig(t, protected, []sdk.AccAddress{victim})
			}
			before := f.balances(victim, recipient, protected)

			err := v090.Sweep(f.ctx, &f.app.AppKeepers, cfg)
			require.ErrorIs(t, err, v090.ErrProtectedAddress, "the sweep must refuse a module account or contract")
			require.Equal(t, before, f.balances(victim, recipient, protected), "a refused sweep moves nothing")
			require.Empty(t, f.sweepEvents(), "a refused sweep emits no sweep event")
		})
	}
}

func instantiateReflectContract(t *testing.T, f sweepFixture) sdk.AccAddress {
	t.Helper()
	code, err := os.ReadFile("../wasmbinding/testdata/reflect.wasm")
	require.NoError(t, err, "the reflect contract is readable")
	creator := sweepAddr(0xCC)
	contracts := wasmkeeper.NewDefaultPermissionKeeper(&f.app.WasmKeeper)
	codeID, _, err := contracts.Create(f.ctx, creator, code, nil)
	require.NoError(t, err, "the reflect contract is stored")
	addr, _, err := contracts.Instantiate(f.ctx, codeID, creator, creator, []byte("{}"), "reflect", nil)
	require.NoError(t, err, "the reflect contract is instantiated")
	return addr
}

func contractAt20Bytes(t *testing.T, f sweepFixture) sdk.AccAddress {
	t.Helper()
	info := f.app.WasmKeeper.GetContractInfo(f.ctx, instantiateReflectContract(t, f))
	require.NotNil(t, info, "the instantiated contract has contract info")
	addr := sweepAddr(0xCD)
	// instantiation yields only 32-byte addresses, which the config refuses, so the contract check is
	// reached by storing a real contract's info under a 20-byte address
	f.ctx.KVStore(f.app.GetKey(wasmtypes.StoreKey)).Set(wasmtypes.GetContractAddressKey(addr), f.app.appCodec.MustMarshal(info))
	require.True(t, f.app.WasmKeeper.HasContractInfo(f.ctx, addr), "the 20-byte address must read as a contract")
	return addr
}

func TestV090ConfigRefusesInstantiatedContract(t *testing.T) {
	f := newSweepFixture(t)
	contract := instantiateReflectContract(t, f)

	_, err := v090.NewChainConfig(sweepAddr(3), []sdk.AccAddress{contract})
	require.ErrorIs(t, err, v090.ErrInvalidAddress, "a 32-byte contract address cannot be listed")
	_, err = v090.NewChainConfig(contract, nil)
	require.ErrorIs(t, err, v090.ErrInvalidAddress, "a 32-byte contract address cannot be the recipient")
}

func TestV090UpgradeHandlerSweepsOnlyConfiguredChains(t *testing.T) {
	// the recipient of the committed lists, made a module account so a configured chain's pre-check refuses it
	pirinRecipient := sdk.MustAccAddressFromBech32("nolus1luvaejqplajhq33xpn42ffsy6558fzg4dvkqnz")

	tests := []struct {
		name    string
		chainID string
		wantErr error
	}{
		{name: "chain without lists runs no sweep", chainID: "unknown-1"},
		{name: "pirin-1 runs its configured sweep", chainID: "pirin-1", wantErr: v090.ErrProtectedAddress},
		{name: "rila-3 runs its configured sweep", chainID: "rila-3", wantErr: v090.ErrProtectedAddress},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newSweepFixture(t)
			macc := authtypes.NewModuleAccount(authtypes.NewBaseAccountWithAddress(pirinRecipient), "recovery-recipient")
			f.app.AccountKeeper.SetAccount(f.ctx, f.app.AccountKeeper.NewAccount(f.ctx, macc))
			f.ctx = f.ctx.WithChainID(tc.chainID)

			handler := v090.CreateUpgradeHandler(f.app.mm, f.app.configurator, &f.app.AppKeepers, f.app.appCodec)
			vm, err := handler(f.ctx, upgradetypes.Plan{Name: v090.UpgradeName}, f.app.mm.GetVersionMap())
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr, "a configured chain's sweep checks that chain's recipient")
			} else {
				require.NoError(t, err, "the v0.9.0 handler succeeds")
				require.Equal(t, f.app.mm.GetVersionMap(), vm, "the handler must return every module at its current consensus version")
			}
			require.Empty(t, f.sweepEvents(), "no address is swept")
		})
	}
}
