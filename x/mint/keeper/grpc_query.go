package keeper

import (
	"context"

	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params := k.GetParams(c)

	return &types.QueryParamsResponse{Params: params}, nil
}

// MintState returns the state minter of the mint module.
func (k Keeper) MintState(c context.Context, _ *types.QueryMintStateRequest) (*types.QueryMintStateResponse, error) {
	minter := k.GetMinter(c)

	return &types.QueryMintStateResponse{NormTimePassed: minter.NormTimePassed, TotalMinted: minter.TotalMinted}, nil
}

// AnnualInflation returns minter.Inflation of the mint module.
func (k Keeper) AnnualInflation(c context.Context, _ *types.QueryAnnualInflationRequest) (*types.QueryAnnualInflationResponse, error) {
	minter := k.GetMinter(c)

	return &types.QueryAnnualInflationResponse{AnnualInflation: minter.AnnualInflation}, nil
}
