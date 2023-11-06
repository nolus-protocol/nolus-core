package types

import (
	"errors"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

var (
	DefaultFeeRate         int32    = 40
	DefaultContractAddress string   = "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz"
	DefaultBaseDenom       string   = sdk.DefaultBondDenom
	DefaultOracleAddress   string   = "nolus1436kxs0w2es6xlqpp9rd35e3d0cjnw4sv8j3a7483sgks29jqwgsv3wzl4"
	DefaultProfitAddress   string   = "nolus1mf6ptkssddfmxvhdx0ech0k03ktp6kf9yk59renau2gvht3nq2gqkxgywu"
	DefaultAcceptedDenoms  []string = []string{"ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y", "ibc/ECFDE61B64BB920E087E7448C4C3FE356B7BD13A1C2153119E98816C964FE196", "unls"}
)

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

// DefaultParams returns default x/tax module parameters.
func DefaultParams() Params {
	return Params{
		FeeRate:         DefaultFeeRate,
		ContractAddress: DefaultContractAddress,
		BaseDenom:       DefaultBaseDenom,
		FeeParams:       DefaultFeeParams(),
	}
}

// DefaultFeeParams is used to initialize the default fee params.
// Oracle and Profit addresses are set to the default addresses which were used in genesis.
func DefaultFeeParams() []*FeeParam {
	return []*FeeParam{
		{
			OracleAddress:  DefaultOracleAddress,
			ProfitAddress:  DefaultProfitAddress,
			AcceptedDenoms: DefaultAcceptedDenoms,
		},
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

	if err := validateFeeParams(p.FeeParams); err != nil {
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
		return errorsmod.Wrap(ErrInvalidAddress, err.Error())
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

func validateFeeParams(v interface{}) error {
	feeParams, ok := v.([]*FeeParam)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	for _, feeParam := range feeParams {
		if feeParam.OracleAddress == "" || feeParam.ProfitAddress == "" ||
			strings.TrimSpace(feeParam.OracleAddress) == "" || strings.TrimSpace(feeParam.ProfitAddress) == "" {
			return ErrInvalidFeeParam
		}
	}

	return nil
}
