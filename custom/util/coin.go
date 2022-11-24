package util

import sdk "github.com/cosmos/cosmos-sdk/types"

func ConvertToMicroNolusInt(amount sdk.Int) sdk.Uint {
	return ConvertToMicroNolusDec(sdk.NewDecFromInt(amount))
}

func ConvertToMicroNolusDec(amount sdk.Dec) sdk.Uint {
	microNolusCoef := sdk.NewDec(10).Power(6)
	return sdk.NewUint(amount.Mul(microNolusCoef).TruncateInt().Uint64())
}
