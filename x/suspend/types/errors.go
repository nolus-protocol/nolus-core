package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/suspend module sentinel errors
var (
	ErrSample = sdkerrors.Register(ModuleName, 1100, "The node is suspended!!!")

)
