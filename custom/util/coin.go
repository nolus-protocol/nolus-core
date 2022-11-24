package util

import sdk "github.com/cosmos/cosmos-sdk/types"

const microNolusCoef = 10000000

func ConvertToMicroNolusInt(amount sdk.Int) sdk.Uint {
	return ConvertToMicroNolusDec(sdk.NewDecFromInt(amount))
}

func ConvertToMicroNolusDec(amount sdk.Dec) sdk.Uint {
	return sdk.NewUint(amount.Mul(sdk.NewDec(microNolusCoef)).TruncateInt().Uint64())
}
