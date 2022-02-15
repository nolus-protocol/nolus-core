package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyFeeRate           = []byte("FeeRate")
	DefaultFeeRate int32 = 40
)

var (
	KeyFeeCaps            = []byte("FeeCaps")
	DefaultFeeCaps string = "1000unolus"
)

var (
	KeyContractAddress            = []byte("ContractAddress")
	DefaultContractAddress string = "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz"
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	feeRate int32,
	feeCaps string,
	contractAddress string,
) Params {
	return Params{
		FeeRate:         feeRate,
		FeeCaps:         feeCaps,
		ContractAddress: contractAddress,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultFeeRate,
		DefaultFeeCaps,
		DefaultContractAddress,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyFeeRate, &p.FeeRate, validateFeeRate),
		paramtypes.NewParamSetPair(KeyFeeCaps, &p.FeeCaps, validateFeeCaps),
		paramtypes.NewParamSetPair(KeyContractAddress, &p.ContractAddress, validateContractAddress),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateFeeRate(p.FeeRate); err != nil {
		return err
	}

	if err := validateFeeCaps(p.FeeCaps); err != nil {
		return err
	}

	if err := validateContractAddress(p.ContractAddress); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateFeeRate validates the FeeRate param
func validateFeeRate(v interface{}) error {
	feeRate, ok := v.(int32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = feeRate

	return nil
}

// validateFeeCaps validates the FeeCaps param
func validateFeeCaps(v interface{}) error {
	feeCaps, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = feeCaps

	return nil
}

// validateContractAddress validates the ContractAddress param
func validateContractAddress(v interface{}) error {
	contractAddress, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = contractAddress

	return nil
}
