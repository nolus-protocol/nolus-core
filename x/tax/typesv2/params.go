package typesv2

import (
	"errors"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultFeeRate                 int32         = 40
	DefaultTreasuryAddress         string        = "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz"
	DefaultBaseDenom               string        = params.BaseCoinUnit
	DefaultProfitAddress           string        = "nolus1mf6ptkssddfmxvhdx0ech0k03ktp6kf9yk59renau2gvht3nq2gqkxgywu"
	DefaultAcceptedDenomsMinPrices []*DenomPrice = []*DenomPrice{
		{
			Denom:    "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y",
			Ticker:   "OSMO",
			MinPrice: "0.025",
		},
		{
			Denom:    "ibc/5DE4FCAF68AE40F81F738C857C0D95F7C1BC47B00FA1026E85C1DD92524D4A11",
			Ticker:   "USDC",
			MinPrice: "0.030",
		},
	}
)

// NewParams creates a new Params instance.
func NewParams(
	feeRate int32,
	treasryAddress string,
	baseDenom string,
) Params {
	return Params{
		FeeRate:         feeRate,
		TreasuryAddress: treasryAddress,
		BaseDenom:       baseDenom,
	}
}

// DefaultParams returns default x/tax module parameters.
func DefaultParams() Params {
	return Params{
		FeeRate:         DefaultFeeRate,
		TreasuryAddress: DefaultTreasuryAddress,
		BaseDenom:       DefaultBaseDenom,
		DexFeeParams:    DefaultFeeParams(),
	}
}

// DefaultFeeParams is used to initialize the default fee params.
// Oracle and Profit addresses are set to the default addresses which were used in genesis.
func DefaultFeeParams() []*DexFeeParams {
	return []*DexFeeParams{
		{
			ProfitAddress:           DefaultProfitAddress,
			AcceptedDenomsMinPrices: DefaultAcceptedDenomsMinPrices,
		},
	}
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if err := validateFeeRate(p.FeeRate); err != nil {
		return err
	}

	if err := validateContractAddress(p.TreasuryAddress); err != nil {
		return err
	}

	if err := validateBaseDenom(p.BaseDenom); err != nil {
		return err
	}

	if err := validateFeeParams(p.DexFeeParams); err != nil {
		return err
	}

	return nil
}

func validateFeeRate(v interface{}) error {
	feeRate, ok := v.(int32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if feeRate < 0 || feeRate > 100 {
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
	feeParams, ok := v.([]*DexFeeParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	for _, feeParam := range feeParams {
		err := validateContractAddress(feeParam.ProfitAddress)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidFeeParam, err.Error())
		}
		if feeParam.ProfitAddress == "" || strings.TrimSpace(feeParam.ProfitAddress) == "" {
			return errorsmod.Wrap(ErrInvalidFeeParam, "profit address cannot be blank")
		}
		for _, denomPrice := range feeParam.AcceptedDenomsMinPrices {
			if denomPrice.Denom == "" || strings.TrimSpace(denomPrice.Denom) == "" ||
				denomPrice.Ticker == "" || strings.TrimSpace(denomPrice.Ticker) == "" {
				return errorsmod.Wrap(ErrInvalidFeeParam, "denom or ticker cannot be blank")
			}
			if denomPrice.MinPrice == "" || strings.TrimSpace(denomPrice.MinPrice) == "" {
				return errorsmod.Wrap(ErrInvalidFeeParam, "min price cannot be blank")
			}
			minPrice, err := sdkmath.LegacyNewDecFromStr(denomPrice.MinPrice)
			if err != nil {
				return errorsmod.Wrap(ErrInvalidFeeParam, err.Error())
			}
			if minPrice.IsZero() || minPrice.IsNegative() {
				return errorsmod.Wrap(ErrInvalidFeeParam, "min price cannot be zero or negative")
			}
		}
	}

	return nil
}
