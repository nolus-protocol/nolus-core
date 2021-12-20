package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	type2 "gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

type NomoSuspendDecorator struct {
	sk SuspendKeeper
}

func NewSuspendDecorator(sk SuspendKeeper) NomoSuspendDecorator {
	return NomoSuspendDecorator{
		sk: sk,
	}
}

func (nsd NomoSuspendDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	state := nsd.sk.GetState(ctx)
	if state.Suspended && (state.BlockHeight == 0 || ctx.BlockHeight() > state.BlockHeight) {
		includesSuspend := false
		for _, msg := range tx.GetMsgs() {
			if _, ok := msg.(*type2.MsgChangeSuspended); ok {
				includesSuspend = true
			}
		}
		if includesSuspend {
			return next(ctx, tx, simulate)
		} else {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "node is suspended")
		}
	}
	return next(ctx, tx, simulate)
}
