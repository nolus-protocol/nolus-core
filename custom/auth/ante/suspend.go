package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	suspendTypes "gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

type NolusSuspendDecorator struct {
	sk SuspendKeeper
}

func NewSuspendDecorator(sk SuspendKeeper) NolusSuspendDecorator {
	return NolusSuspendDecorator{
		sk: sk,
	}
}

func (nsd NolusSuspendDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	state := nsd.sk.GetState(ctx)
	if state.Suspended && ctx.BlockHeight() > state.BlockHeight {
		includesSuspend := false
		for _, msg := range tx.GetMsgs() {
			if _, ok := msg.(*suspendTypes.MsgChangeSuspended); ok {
				includesSuspend = true
			}
		}
		if includesSuspend {
			return next(ctx, tx, simulate)
		} else {
			return ctx, sdkerrors.Wrap(suspendTypes.ErrSuspended, "unauthorized")
		}
	}
	return next(ctx, tx, simulate)
}
