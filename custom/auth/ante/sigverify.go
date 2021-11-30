package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

///////////
type NomoSuspendDecorator struct {
	tk SuspendKeeper
}
func NewSuspendDecorator(tk SuspendKeeper) NomoSuspendDecorator {
	return NomoSuspendDecorator{
		tk: tk,
	}
}

func (mfd NomoSuspendDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if mfd.tk.IsNodeSuspend() {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "The node is suspended!")
	}
	return next(ctx, tx, simulate)
}
