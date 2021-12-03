package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)


func (k msgServer) ChangeSuspend(goCtx context.Context,  msg *types.MsgChangeSuspend) (*types.MsgChangeSuspendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.SetNodeSuspend(ctx, msg)
    // TODO: Handling the message
    _ = ctx

	return &types.MsgChangeSuspendResponse{}, nil
}
