package typesv2

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tax module sentinel errors.
var (
	ErrInvalidFeeRate  = errorsmod.Register(ModuleName, 1, "feeRate should be between 0 and 100")
	ErrInvalidAddress  = errorsmod.Register(ModuleName, 2, "invalid address")
	ErrTooManyFeeCoins = errorsmod.Register(ModuleName, 3, "only one fee denom per tx")
	ErrInvalidFeeDenom = errorsmod.Register(ModuleName, 4, "denom is not allowed")
	ErrInvalidTax      = errorsmod.Register(ModuleName, 6, "tax can not be negative, zero or nil")
	ErrInvalidFeeParam = errorsmod.Register(ModuleName, 7, "current fee param is not valid")
	ErrNoPrices        = errorsmod.Register(ModuleName, 8, "no prices found from the oracle")
)
