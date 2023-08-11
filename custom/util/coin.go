package util

import "cosmossdk.io/math"

const microNolusCoef = 1000000

func ConvertToMicroNolusInt64(amount int64) math.Uint {
	return ConvertToMicroNolusDec(math.LegacyNewDec(amount))
}

func ConvertToMicroNolusDec(amount math.LegacyDec) math.Uint {
	return math.NewUint(amount.Mul(math.LegacyNewDec(microNolusCoef)).TruncateInt().Uint64())
}
