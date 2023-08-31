package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Nolus-Protocol/nolus-core/custom/util"
)

// Minting formula f(x)=-4.33275 x^3 + 944.61206 x^2 - 88567.25194 x + 3.86335×10^6 integrated over 0.47 to 96
// afterwards minting 103125 tokens each month until reaching the minting cap of 150*10^6 tokens.
var (
	QuadCoef          = sdkmath.LegacyMustNewDecFromStr("-1.08319")
	CubeCoef          = sdkmath.LegacyMustNewDecFromStr("314.871")
	SquareCoef        = sdkmath.LegacyMustNewDecFromStr("-44283.6")
	Coef              = sdkmath.LegacyMustNewDecFromStr("3863350")
	MintingCap        = util.ConvertToMicroNolusInt64(150000000)
	FixedMintedAmount = util.ConvertToMicroNolusInt64(103125)
	NormOffset        = sdkmath.LegacyMustNewDecFromStr("0.47")
	MonthsInFormula   = sdkmath.LegacyMustNewDecFromStr("96")
	TotalMonths       = sdkmath.LegacyMustNewDecFromStr("120")
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

	if minter.TotalMinted.GT(MintingCap) {
		return fmt.Errorf("mint parameter totalMinted: %v can not be bigger than MintingCap: %v",
			minter.TotalMinted, MintingCap)
	}

	calculatedMintedTokens := calcMintedTokens(minter)

	if minter.NormTimePassed.GT(TotalMonths.Sub(sdkmath.LegacyNewDec(1))) {
		if calculatedMintedTokens.GT(MintingCap) || MintingCap.Sub(calculatedMintedTokens).GT(FixedMintedAmount) {
			return fmt.Errorf("mint parameters are not conformant with the minting schedule, for %s month minted %s unls",
				minter.NormTimePassed, calculatedMintedTokens)
		}
	} else if !calculatedMintedTokens.Equal(minter.TotalMinted) {
		return fmt.Errorf("minted unexpected amount of tokens for %s months. act: %v, exp: %v",
			minter.NormTimePassed, minter.TotalMinted, calculatedMintedTokens)
	}

	return nil
}

func calcMintedTokens(m Minter) sdkmath.Uint {
	if m.NormTimePassed.GTE(MonthsInFormula) {
		fixedMonthsPeriod := sdkmath.NewUint(m.NormTimePassed.Sub(MonthsInFormula).TruncateInt().Uint64())
		fixedMonthsTokens := fixedMonthsPeriod.Mul(FixedMintedAmount)
		calculatedTokensByIntegral := CalcTokensByIntegral(MonthsInFormula).Sub(CalcTokensByIntegral(NormOffset))

		return calculatedTokensByIntegral.Add(fixedMonthsTokens)
	} else {
		return CalcTokensByIntegral(m.NormTimePassed).Sub(CalcTokensByIntegral(NormOffset))
	}
}

// Integral:  -1.08319 x^4 + 314.871 x^3 - 44283.6 x^2 + 3.86335×10^6 x
// transformed to: (((-1.08319 x + 314.871) x - 44283.6) x +3.86335×10^6) x.
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
