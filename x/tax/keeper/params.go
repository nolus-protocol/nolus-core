package keeper

import (
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO add storeService storetypes.KVStoreService as a field in the keeper and use context.context here;
// refer to the cosmos-sdk SetParams() to see how they reach the KVStore -> store := k.storeService.OpenKVStore(ctx)
// OR use unwrapSDKContext / do this for all of our custom modules

// GetParams get all parameters as types.Params.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams set the params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	return nil
}

// FeeRate returns the fee rate.
func (k Keeper) FeeRate(ctx sdk.Context) (res int32) {
	var p types.Params
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return 0
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p.FeeRate
}

// ContractAddress returns the contract address.
func (k Keeper) ContractAddress(ctx sdk.Context) (res string) {
	var p types.Params
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return ""
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p.ContractAddress
}

// BseDenom returns the base denom.
func (k Keeper) BaseDenom(ctx sdk.Context) (res string) {
	var p types.Params
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return ""
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p.BaseDenom
}
