package keeper

import (
	"context"

	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
)

// GetParams get all parameters as types.Params.
func (k Keeper) GetParams(ctx context.Context) types.Params {
	store := k.storeService.OpenKVStore(ctx)

	var p types.Params
	b, err := store.Get(types.ParamsKey)
	if err != nil {
		// TODO panic("error getting stored tax params")
		return p
	}

	k.cdc.MustUnmarshal(b, &p)
	return p
}

// SetParams set the params.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&params)

	if err := store.Set(types.ParamsKey, b); err != nil {
		return err
	}

	return nil
}

// FeeRate returns the fee rate.
func (k Keeper) FeeRate(ctx context.Context) int32 {
	store := k.storeService.OpenKVStore(ctx)

	b, err := store.Get(types.ParamsKey)
	if err != nil {
		return 0
	}

	var p types.Params
	k.cdc.MustUnmarshal(b, &p)
	return p.FeeRate
}

// ContractAddress returns the contract address.
func (k Keeper) ContractAddress(ctx context.Context) string {
	store := k.storeService.OpenKVStore(ctx)

	b, err := store.Get(types.ParamsKey)
	if err != nil {
		return ""
	}

	var p types.Params
	k.cdc.MustUnmarshal(b, &p)
	return p.ContractAddress
}

// BseDenom returns the base denom.
func (k Keeper) BaseDenom(ctx context.Context) string {
	store := k.storeService.OpenKVStore(ctx)

	b, err := store.Get(types.ParamsKey)
	if err != nil {
		return ""
	}

	var p types.Params
	k.cdc.MustUnmarshal(b, &p)
	return p.BaseDenom
}
