package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

func (k msgServer) ChangeSuspended(goCtx context.Context, msg *types.MsgChangeSuspended) (*types.MsgChangeSuspendedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.ChangeSuspendedState(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgChangeSuspendedResponse{}, nil
}
