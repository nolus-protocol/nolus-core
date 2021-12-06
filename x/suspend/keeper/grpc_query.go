package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

var _ types.QueryServer = Keeper{}

// Supend returns params of the mint module.
func (k Keeper) Suspend(c context.Context, _ *types.QuerySuspendRequest) (*types.QuerySuspendResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.IsNodeSuspend(ctx)
	qs := types.QuerySuspend{
		Creator: params.Creator,
		Suspend: params.Suspend,
	}

	return &types.QuerySuspendResponse{QuerySuspend: &qs}, nil
}
