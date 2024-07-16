package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

// Keeper of the mint store.
type Keeper struct {
	cdc              codec.BinaryCodec
	storeService     store.KVStoreService
	bankKeeper       types.BankKeeper
	feeCollectorName string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates a new mint Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService,
	ak types.AccountKeeper, bk types.BankKeeper,
	feeCollectorName string, authority string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("the x/%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		cdc:              cdc,
		storeService:     storeService,
		bankKeeper:       bk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
	}
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	c := sdk.UnwrapSDKContext(ctx)
	return c.Logger().With("module", "x/"+types.ModuleName)
}

// GetMinter get the minter.
func (k Keeper) GetMinter(ctx context.Context) types.Minter {
	store := k.storeService.OpenKVStore(ctx)

	b, err := store.Get(types.MinterKey)
	if err != nil {
		panic("error getting stored minter")
	}

	var minter types.Minter
	k.cdc.MustUnmarshal(b, &minter)
	return minter
}

// SetMinter set the minter.
func (k Keeper) SetMinter(ctx context.Context, minter types.Minter) error {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&minter)

	err := store.Set(types.MinterKey, b)
	if err != nil {
		return err
	}

	return nil
}

// GetParams returns the current x/mint module parameters.
func (k Keeper) GetParams(ctx context.Context) types.Params {
	store := k.storeService.OpenKVStore(ctx)

	b, err := store.Get(types.ParamsKey)
	if err != nil {
		panic("error getting stored minter")
	}

	var params types.Params
	k.cdc.MustUnmarshal(b, &params)
	return params
}

// SetParams sets the x/mint module parameters.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&p)

	if err := store.Set(types.ParamsKey, b); err != nil {
		return err
	}

	return nil
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx context.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in BeginBlocker.
func (k Keeper) AddCollectedFees(ctx context.Context, fees sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}
