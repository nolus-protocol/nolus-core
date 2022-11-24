package util

import sdk "github.com/cosmos/cosmos-sdk/types"

var microNolusCoef = sdk.NewDec(10).Power(6)

func ConvertToMicroNolusInt(amount sdk.Int) sdk.Uint {
	return ConvertToMicroNolusDec(sdk.NewDecFromInt(amount))
}

func ConvertToMicroNolusDec(amount sdk.Dec) sdk.Uint {
	return sdk.NewUint(amount.Mul(microNolusCoef).TruncateInt().Uint64())
}
