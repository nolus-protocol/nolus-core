package keeper

import (
	"context"

	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
)

// GetParams get all parameters as types.Params.
func (k Keeper) GetParams(ctx context.Context) (params types.Params, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return params, err
	}
	if bz == nil {
		return params, nil
	}

	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}

// SetParams set the params.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	return store.Set(types.ParamsKey, bz)
}

// FeeRate returns the fee rate.
func (k Keeper) FeeRate(ctx context.Context) int32 {
	p, err := k.GetParams(ctx)
	if err != nil {
		return 0
	}
	return p.FeeRate
}

// TreasuryAddress returns the contract address.
func (k Keeper) TreasuryAddress(ctx context.Context) string {
	p, err := k.GetParams(ctx)
	if err != nil {
		return ""
	}
	return p.TreasuryAddress
}

// BaseDenom returns the base denom.
func (k Keeper) BaseDenom(ctx context.Context) string {
	p, err := k.GetParams(ctx)
	if err != nil {
		return ""
	}
	return p.BaseDenom
}
