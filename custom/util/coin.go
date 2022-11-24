package util

import sdk "github.com/cosmos/cosmos-sdk/types"

const microNolusCoef = 10000000

func ConvertToMicroNolusInt(amount int64) sdk.Uint {
	return ConvertToMicroNolusDec(sdk.NewDecFromInt(sdk.NewInt(amount)))
}

func ConvertToMicroNolusDec(amount sdk.Dec) sdk.Uint {
	return sdk.NewUint(amount.Mul(sdk.NewDec(microNolusCoef)).TruncateInt().Uint64())
}
