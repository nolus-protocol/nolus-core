package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Nolus-Protocol/nolus-core/custom/util"
)

// Legacy Minting formula integral-4.33275 x^3 + 952.82456 x^2 - 88567.49981 x + 3.86381×10^6 dx = -1.08319 x^4 + 317.608 x^3 - 44283.7 x^2 + 3863810 x + constant

// Current Minting formula integral-0.11175 x^3 + 50.82456 x^2 - 1767.49981 x + 0.83381×10^6 dx = -0.0279375 x^4 + 16.9415 x^3 - 883.75 x^2 + 833810 x + constant
var (
	QuadCoef          = sdkmath.LegacyMustNewDecFromStr("-0.0279375")
	CubeCoef          = sdkmath.LegacyMustNewDecFromStr("16.9415")
	SquareCoef        = sdkmath.LegacyMustNewDecFromStr("-883.75")
	Coef              = sdkmath.LegacyMustNewDecFromStr("833810")
	FixedMintedAmount = util.ConvertToMicroNolusInt64(103125)
	NormOffset        = sdkmath.LegacyMustNewDecFromStr("0.17")
	TotalMonths       = sdkmath.LegacyMustNewDecFromStr("120")
	MonthsInFormula   = TotalMonths
	AbsMonthsRange    = MonthsInFormula.Sub(NormOffset)
	NormMonthsRange   = AbsMonthsRange.Quo(MonthsInFormula)
)

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(normTimePassed sdkmath.LegacyDec, totalMinted, prevBlockTimestamp, inflation sdkmath.Uint) Minter {
	return Minter{
		NormTimePassed:     normTimePassed,
		TotalMinted:        totalMinted,
		PrevBlockTimestamp: prevBlockTimestamp,
		AnnualInflation:    inflation,
	}
}

// InitialMinter returns an initial Minter object with zero-value parameters.
func InitialMinter() Minter {
	return NewMinter(
		NormOffset,
		sdkmath.ZeroUint(),
		sdkmath.ZeroUint(),
		sdkmath.ZeroUint(),
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain.
func DefaultInitialMinter() Minter {
	return InitialMinter()
}

// ValidateMinter ensure minter has valid "normTimePassed" and
// "totalMinted" tokens do not exceed the minting cap.
func ValidateMinter(minter Minter) error {
	if minter.NormTimePassed.IsNegative() {
		return fmt.Errorf("mint parameter normTimePassed should be positive, is %s",
			minter.NormTimePassed.String())
	}

	if minter.NormTimePassed.GT(TotalMonths) {
		return fmt.Errorf("mint parameter normTimePassed: %v should not be bigger than TotalMonths: %v", minter.NormTimePassed, TotalMonths)
	}

	calculatedMintedTokens := calcMintedTokens(minter)

	if !calculatedMintedTokens.Equal(minter.TotalMinted) {
		return fmt.Errorf("minted unexpected amount of tokens for %s months. act: %v, exp: %v",
			minter.NormTimePassed, minter.TotalMinted, calculatedMintedTokens)
	}

	return nil
}

func calcMintedTokens(m Minter) sdkmath.Uint {
	return CalcTokensByIntegral(m.NormTimePassed).Sub(CalcTokensByIntegral(NormOffset))
}

// Integral:  -0.0279375 x^4 + 16.9415 x^3 - 883.75 x^2 + 833810 x + constant
// transformed to: (((-0.0279375 x + 16.9415) x - 883.75) x + 833810) x.
func CalcTokensByIntegral(x sdkmath.LegacyDec) sdkmath.Uint {
	return util.ConvertToMicroNolusDec(((((QuadCoef.Mul(x).Add(CubeCoef)).Mul(x).Add(SquareCoef)).Mul(x).Add(Coef)).Mul(x)))
}

func GetAbsDiff(a, b sdkmath.Uint) sdkmath.Uint {
	if a.GTE(b) {
		return a.Sub(b)
	}

	return b.Sub(a)
}

func DecFromUint(u sdkmath.Uint) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecFromBigInt(u.BigInt())
}
