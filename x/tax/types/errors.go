package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tax module sentinel errors
var (
	ErrInvalidFeeRate = sdkerrors.Register(ModuleName, 1, "feeRate should be between 0 and 100")
	ErrInvalidAddress = sdkerrors.Register(ModuleName, 2, "invalid address")
)
