package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tax module sentinel errors
var (
	ErrInvalidFeeRate    = sdkerrors.Register(ModuleName, 1, "feeRate should be between 0 and 100")
	ErrInvalidAddress    = sdkerrors.Register(ModuleName, 2, "invalid address")
	ErrDuplicateFeeDenom = sdkerrors.Register(ModuleName, 3, "duplicate fee denoms are not allowed")
	ErrTooManyFeeCoins   = sdkerrors.Register(ModuleName, 4, "only one fee denom per tx")
	ErrInvalidFeeDenom   = sdkerrors.Register(ModuleName, 5, "denom is not allowed")
	ErrAmountNilOrZero   = sdkerrors.Register(ModuleName, 6, "amount can not be nil or zero")
	ErrFeesNotSet        = sdkerrors.Register(ModuleName, 7, "fees can not be empty")
)
