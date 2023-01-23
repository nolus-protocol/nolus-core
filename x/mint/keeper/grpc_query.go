package keeper

import (
	"context"

	"github.com/Nolus-Protocol/nolus-core/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// MintState returns the state minter of the mint module.
func (k Keeper) MintState(c context.Context, _ *types.QueryMintStateRequest) (*types.QueryMintStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)

	return &types.QueryMintStateResponse{NormTimePassed: minter.NormTimePassed, TotalMinted: minter.TotalMinted}, nil
}

// AnnualInflation returns minter.Inflation of the mint module.
func (k Keeper) AnnualInflation(c context.Context, _ *types.QueryAnnualInflationRequest) (*types.QueryAnnualInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)

	return &types.QueryAnnualInflationResponse{AnnualInflation: minter.AnnualInflation}, nil
}
