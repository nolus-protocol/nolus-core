package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tax module sentinel errors.
var (
	ErrInvalidFeeRate  = sdkerrors.Register(ModuleName, 1, "feeRate should be between 0 and 50")
	ErrInvalidAddress  = sdkerrors.Register(ModuleName, 2, "invalid address")
	ErrTooManyFeeCoins = sdkerrors.Register(ModuleName, 3, "only one fee denom per tx")
	ErrInvalidFeeDenom = sdkerrors.Register(ModuleName, 4, "denom is not allowed")
	ErrAmountNilOrZero = sdkerrors.Register(ModuleName, 5, "amount can not be nil or zero")
	ErrInvalidTax      = sdkerrors.Register(ModuleName, 6, "tax can not be negative, zero or nil")
)
