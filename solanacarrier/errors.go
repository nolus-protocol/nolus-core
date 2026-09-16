package solanacarrier

import errorsmod "cosmossdk.io/errors"

// x/solanacarrier module sentinel errors.
var (
	ErrInvalidFeePayer = errorsmod.Register(ModuleName, 1, "invalid fee payer")
	ErrParamsNotFound  = errorsmod.Register(ModuleName, 2, "solanacarrier params not found")
	ErrNoSDKContext    = errorsmod.Register(ModuleName, 3, "no sdk context attached")
)
