package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

var _ types.QueryServer = Keeper{}

// SuspendedState returns the state of suspend module.
func (k Keeper) SuspendedState(c context.Context, _ *types.QuerySuspendRequest) (*types.QuerySuspendResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	state := k.GetState(ctx)
	qs := types.QuerySuspendResponse{
		State: state,
	}
	return &qs, nil
}
