package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.FeeRate(ctx),
		k.ContractAddress(ctx),
		k.BaseDenom(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// FeeRate returns the fee rate
func (k Keeper) FeeRate(ctx sdk.Context) (res int32) {
	k.paramstore.Get(ctx, types.KeyFeeRate, &res)
	return
}

// ContractAddress returns the contract address
func (k Keeper) ContractAddress(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyContractAddress, &res)
	return
}

// BseDenom returns the base denom
func (k Keeper) BaseDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyBaseDenom, &res)
	return
}
