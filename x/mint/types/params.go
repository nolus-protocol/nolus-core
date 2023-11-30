package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Parameter store keys.
var (
	DefaultBondDenom              = sdk.DefaultBondDenom
	DefaultMaxMintablenanoseconds = int64(time.Minute) // 1 minute default
)

func NewParams(mintDenom string, maxMintableNanoseconds sdkmath.Uint) Params {
	return Params{
		MintDenom:              mintDenom,
		MaxMintableNanoseconds: maxMintableNanoseconds,
	}
}

// DefaultParams returns default x/mint module parameters.
func DefaultParams() Params {
	return Params{
		MintDenom:              sdk.DefaultBondDenom,
		MaxMintableNanoseconds: sdkmath.NewUint(60000000000), // 1 minute default
	}
}

// Validate validate params.
func (p Params) Validate() error {
	if err := validateMaxMintableNanoseconds(p.MaxMintableNanoseconds); err != nil {
		return err
	}
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}

	return nil
}

func validateMaxMintableNanoseconds(i interface{}) error {
	v, ok := i.(sdkmath.Uint)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LTE(sdkmath.ZeroUint()) {
		return fmt.Errorf("max mintable period must be positive: %d", v)
	}

	return nil
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}
