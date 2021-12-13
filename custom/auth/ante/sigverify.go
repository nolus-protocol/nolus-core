package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	type2 "gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

type NomoSuspendDecorator struct {
	tk SuspendKeeper
}

func NewSuspendDecorator(tk SuspendKeeper) NomoSuspendDecorator {
	return NomoSuspendDecorator{
		tk: tk,
	}
}

func (mfd NomoSuspendDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgSuspend := mfd.tk.IsNodeSuspend(ctx)

	if msgSuspend.Suspend {
		includeSuspend := false
		for _, msg := range tx.GetMsgs() {
			if _, ok := msg.(*type2.MsgChangeSuspend); ok {
				includeSuspend = true
			}
		}
		if includeSuspend {
			return next(ctx, tx, simulate)
		}
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "The node is suspended!")
	}
	return next(ctx, tx, simulate)
}
