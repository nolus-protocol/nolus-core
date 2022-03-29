package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

type SuspendDecorator struct {
	sk Keeper
}

func NewSuspendDecorator(sk Keeper) SuspendDecorator {
	return SuspendDecorator{
		sk: sk,
	}
}

func (nsd SuspendDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	state := nsd.sk.GetState(ctx)
	if state.Suspended && ctx.BlockHeight() > state.BlockHeight {
		includesUnsuspend := false
		for _, msg := range tx.GetMsgs() {
			if _, ok := msg.(*types.MsgUnsuspend); ok {
				includesUnsuspend = true
			}
		}
		if includesUnsuspend {
			return next(ctx, tx, simulate)
		} else {
			return ctx, sdkerrors.Wrap(types.ErrSuspended, "unauthorized")
		}
	}
	return next(ctx, tx, simulate)
}
