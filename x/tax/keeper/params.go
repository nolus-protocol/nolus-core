package keeper

import (
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.FeeRate(ctx),
		k.FeeCaps(ctx),
		k.ContractAddress(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}


// FeeRate returns the FeeRate param
func (k Keeper) FeeRate(ctx sdk.Context) (res int32) {
	k.paramstore.Get(ctx, types.KeyFeeRate, &res)
	return
}

// FeeCaps returns the FeeCaps param
func (k Keeper) FeeCaps(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyFeeCaps, &res)
	return
}

// ContractAddress returns the ContractAddress param
func (k Keeper) ContractAddress(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyContractAddress, &res)
	return
}
