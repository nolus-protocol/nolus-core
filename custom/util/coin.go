package util

import sdk "github.com/cosmos/cosmos-sdk/types"

const microNolusCoef = 1000000

func ConvertToMicroNolusInt64(amount int64) sdk.Uint {
	return ConvertToMicroNolusDec(sdk.NewDec(amount))
}

func ConvertToMicroNolusDec(amount sdk.Dec) sdk.Uint {
	return sdk.NewUint(amount.Mul(sdk.NewDec(microNolusCoef)).TruncateInt().Uint64())
}
