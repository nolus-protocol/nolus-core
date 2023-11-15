package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tax module sentinel errors.
var (
	ErrInvalidFeeRate  = errorsmod.Register(ModuleName, 1, "feeRate should be between 0 and 50")
	ErrInvalidAddress  = errorsmod.Register(ModuleName, 2, "invalid address")
	ErrTooManyFeeCoins = errorsmod.Register(ModuleName, 3, "only one fee denom per tx")
	ErrInvalidFeeDenom = errorsmod.Register(ModuleName, 4, "denom is not allowed")
	ErrAmountNilOrZero = errorsmod.Register(ModuleName, 5, "amount can not be nil or zero")
	ErrInvalidTax      = errorsmod.Register(ModuleName, 6, "tax can not be negative, zero or nil")
	ErrInvalidFeeParam = errorsmod.Register(ModuleName, 7, "current fee param is not valid")
)
