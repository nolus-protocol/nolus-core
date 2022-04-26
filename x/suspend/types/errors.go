package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/suspend module sentinel errors
var (
	ErrSuspended = sdkerrors.Register(ModuleName, 1100, "node is suspended")
)
