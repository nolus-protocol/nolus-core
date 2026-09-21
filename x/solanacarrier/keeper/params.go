package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// GetParams gets all parameters as types.Params.
func (k Keeper) GetParams(ctx context.Context) (params types.Params, err error) {
	// The store service unwraps the sdk.Context and panics when there is none.
	// FeePayer is reached from the sign-mode handler, whose context is not
	// guaranteed to carry one, so the absence has to be an error, not a panic.
	if err := assertSDKContext(ctx); err != nil {
		return params, err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return params, err
	}
	if bz == nil {
		return params, types.ErrParamsNotFound
	}

	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}

// SetParams sets the params.
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

// FeePayer returns the account key every carrier must place at account index 0.
// It satisfies solanacarrier.FeePayerSource, so an unreadable or unset parameter
// must surface as an error rather than an empty key the sign-mode handler would
// silently match against.
func (k Keeper) FeePayer(ctx context.Context) ([]byte, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return params.FeePayer, nil
}

// assertSDKContext mirrors sdk.UnwrapSDKContext's lookup without its panic.
func assertSDKContext(ctx context.Context) error {
	if _, ok := ctx.(sdk.Context); ok {
		return nil
	}
	if _, ok := ctx.Value(sdk.SdkContextKey).(sdk.Context); ok {
		return nil
	}

	return types.ErrNoSDKContext
}
