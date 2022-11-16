package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/custom/util"
)

// Minting formula f(x)=-4.33275 x^3 + 944.61206 x^2 - 88567.25194 x + 3.86335×10^6 integrated over 0.47 to 96
// afterwards minting 103125 tokens each month until reaching the minting cap of 150*10^6 tokens
var (
	QuadCoef          = sdk.MustNewDecFromStr("-1.08319")
	CubeCoef          = sdk.MustNewDecFromStr("314.871")
	SquareCoef        = sdk.MustNewDecFromStr("-44283.6")
	Coef              = sdk.MustNewDecFromStr("3863350")
	MintingCap        = util.ConvertToMicroNolusInt(sdk.NewInt(150000000))
	FixedMintedAmount = util.ConvertToMicroNolusInt(sdk.NewInt(103125))
	NormOffset        = sdk.MustNewDecFromStr("0.47")
	MonthsInFormula   = sdk.MustNewDecFromStr("96")
	TotalMonths       = sdk.MustNewDecFromStr("120")
	AbsMonthsRange    = MonthsInFormula.Sub(NormOffset)
	NormMonthsRange   = AbsMonthsRange.Quo(MonthsInFormula)
)

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(normTimePassed sdk.Dec, totalMinted sdk.Int, prevBlockTimestamp int64) Minter {
	return Minter{
		NormTimePassed:     normTimePassed,
		TotalMinted:        totalMinted,
		PrevBlockTimestamp: prevBlockTimestamp,
	}
}

// InitialMinter returns an initial Minter object with zero-value parameters.
func InitialMinter() Minter {
	return NewMinter(
		NormOffset,
		sdk.NewInt(0),
		int64(0),
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain.
func DefaultInitialMinter() Minter {
	return InitialMinter()
}

// ValidateMinter validate minter.
func ValidateMinter(minter Minter) error {
	if minter.NormTimePassed.IsNegative() {
		return fmt.Errorf("mint parameter normTimePassed should be positive, is %s",
			minter.NormTimePassed.String())
	}
	if minter.TotalMinted.IsNegative() {
		return fmt.Errorf("mint parameter totalMinted should be positive, is %s",
			minter.TotalMinted.String())
	}
	if minter.TotalMinted.GT(MintingCap) {
		return fmt.Errorf("mint parameter totalMinted can not be bigger than MintingCap, is %s",
			minter.TotalMinted)
	}
	calculatedTotalTokens := calcTotalTokens(minter)
	if minter.NormTimePassed.Equal(MonthsInFormula) && minter.TotalMinted.GT(MintingCap) {
		if (MintingCap.Sub(minter.TotalMinted)).GT(FixedMintedAmount) {
			return fmt.Errorf("mint parameters are not conformant with the minting schedule, for %s month minted %s unls",
				minter.NormTimePassed, minter.TotalMinted)
		}
	}
	if !calculatedTotalTokens.Equal(minter.TotalMinted) {
		return fmt.Errorf("minted unexpected ammount of tokens for %s months: %s unls",
			minter.NormTimePassed, minter.TotalMinted)
	}

	return nil
}

func calcTotalTokens(m Minter) sdk.Int {
	calculatedTokensByIntegral := CalcTokensByIntegral(m.NormTimePassed).Sub(CalcTokensByIntegral(NormOffset))
	if m.NormTimePassed.GT(MonthsInFormula) {
		return calculatedTokensByIntegral.Add((util.ConvertToMicroNolusDec(m.NormTimePassed.Sub(MonthsInFormula))).Mul(FixedMintedAmount))
	} else {
		return calculatedTokensByIntegral
	}
}

// Integral:  -1.08319 x^4 + 314.871 x^3 - 44283.6 x^2 + 3.86335×10^6 x
// transformed to: (((-1.08319 x + 314.871) x - 44283.6) x +3.86335×10^6) x
func CalcTokensByIntegral(x sdk.Dec) sdk.Int {
	return util.ConvertToMicroNolusDec(((((QuadCoef.Mul(x).Add(CubeCoef)).Mul(x).Add(SquareCoef)).Mul(x).Add(Coef)).Mul(x)))
}
