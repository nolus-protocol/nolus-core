package util

import sdk "github.com/cosmos/cosmos-sdk/types"

func ConvertToMicroNolusInt(amount sdk.Int) sdk.Int {
	return ConvertToMicroNolusDec(sdk.NewDecFromInt(amount))
}

func ConvertToMicroNolusDec(amount sdk.Dec) sdk.Int {
	return amount.Mul(sdk.NewDec(10).Power(6)).TruncateInt()
}
