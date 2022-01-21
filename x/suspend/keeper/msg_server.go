package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) Suspend(goCtx context.Context, msg *types.MsgSuspend) (*types.MsgSuspendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.SetSuspendState(ctx, true, msg.FromAddress, msg.BlockHeight)
	if err != nil {
		return nil, err
	}

	return &types.MsgSuspendResponse{}, nil
}

func (k msgServer) Unsuspend(goCtx context.Context, msg *types.MsgUnsuspend) (*types.MsgUnsuspendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.SetSuspendState(ctx, false, msg.FromAddress, 0)
	if err != nil {
		return nil, err
	}

	return &types.MsgUnsuspendResponse{}, nil
}
