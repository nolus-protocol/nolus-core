package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var (
	KeyFeeRate           = []byte("FeeRate")
	DefaultFeeRate int32 = 40

	KeyContractAddress            = []byte("ContractAddress")
	DefaultContractAddress string = "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz"

	KeyBaseDenom            = []byte("BaseDenom")
	DefaultBaseDenom string = sdk.DefaultBondDenom
)

// ParamKeyTable the param key table for launch module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance.
func NewParams(
	feeRate int32,
	contractAddress string,
	baseDenom string,
) Params {
	return Params{
		FeeRate:         feeRate,
		ContractAddress: contractAddress,
		BaseDenom:       baseDenom,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultFeeRate,
		DefaultContractAddress,
		DefaultBaseDenom,
	)
}

// ParamSetPairs get the params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyFeeRate, &p.FeeRate, validateFeeRate),
		paramtypes.NewParamSetPair(KeyContractAddress, &p.ContractAddress, validateContractAddress),
		paramtypes.NewParamSetPair(KeyBaseDenom, &p.BaseDenom, validateBaseDenom),
	}
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if err := validateFeeRate(p.FeeRate); err != nil {
		return err
	}

	if err := validateContractAddress(p.ContractAddress); err != nil {
		return err
	}

	if err := validateBaseDenom(p.BaseDenom); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateFeeRate(v interface{}) error {
	feeRate, ok := v.(int32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if feeRate < 0 || feeRate > 50 {
		return ErrInvalidFeeRate
	}

	return nil
}

func validateContractAddress(v interface{}) error {
	contractAddress, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	_, err := sdk.AccAddressFromBech32(contractAddress)
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidAddress, err.Error())
	}

	return nil
}

func validateBaseDenom(v interface{}) error {
	baseDenom, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if strings.TrimSpace(baseDenom) == "" {
		return errors.New("base denom cannot be blank")
	}

	err := sdk.ValidateDenom(baseDenom)
	if err != nil {
		return err
	}

	return nil
}
