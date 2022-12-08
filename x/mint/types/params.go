package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyMintDenom     = []byte("MintDenom")
	DefaultBondDenom = sdk.DefaultBondDenom

	KeyMaxMintableNanoseconds     = []byte("MaxMintableNanoseconds")
	DefaultMaxMintablenanoseconds = int64(time.Minute) // 1 minute default
)

// ParamKeyTable ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(mintDenom string, maxMintableNanoseconds sdk.Uint) Params {
	return Params{
		MintDenom:              mintDenom,
		MaxMintableNanoseconds: maxMintableNanoseconds,
	}
}

// DefaultParams default minting module parameters.
func DefaultParams() Params {
	return Params{
		MintDenom:              sdk.DefaultBondDenom,
		MaxMintableNanoseconds: sdk.NewUint(60000000000), // 1 minute default
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

// ParamSetPairs Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxMintableNanoseconds, &p.MaxMintableNanoseconds, validateMaxMintableNanoseconds),
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
	}
}

func validateMaxMintableNanoseconds(i interface{}) error {
	v, ok := i.(sdk.Uint)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LTE(sdk.ZeroUint()) {
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
