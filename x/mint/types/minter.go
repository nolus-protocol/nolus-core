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
	MintingCap        = util.ConvertToMicroNolusInt64(150000000)
	FixedMintedAmount = util.ConvertToMicroNolusInt64(103125)
	NormOffset        = sdk.MustNewDecFromStr("0.47")
	MonthsInFormula   = sdk.MustNewDecFromStr("96")
	TotalMonths       = sdk.MustNewDecFromStr("120")
	AbsMonthsRange    = MonthsInFormula.Sub(NormOffset)
	NormMonthsRange   = AbsMonthsRange.Quo(MonthsInFormula)
)

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(normTimePassed sdk.Dec, totalMinted sdk.Uint, prevBlockTimestamp sdk.Uint) Minter {
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
		sdk.NewUint(0),
		sdk.NewUint(0),
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
	if minter.NormTimePassed.GT(TotalMonths) {
		return fmt.Errorf("mint parameter normTimePassed: %v should not be bigger than TotalMonths: %v", minter.NormTimePassed, TotalMonths)
	}
	if minter.TotalMinted.GT(MintingCap) {
		return fmt.Errorf("mint parameter totalMinted can not be bigger than MintingCap, is %s",
			minter.TotalMinted)
	}
	calculatedMintedTokens := calcMintedTokens(minter)
	if minter.NormTimePassed.GT(TotalMonths.Sub(sdk.NewDec(1))) {
		if calculatedMintedTokens.GT(MintingCap) || MintingCap.Sub(calculatedMintedTokens).GT(FixedMintedAmount) {
			return fmt.Errorf("mint parameters are not conformant with the minting schedule, for %s month minted %s unls",
				minter.NormTimePassed, calculatedMintedTokens)
		}
	} else if !calculatedMintedTokens.Equal(minter.TotalMinted) {
		return fmt.Errorf("minted unexpected ammount of tokens for %s months: %s unls",
			minter.NormTimePassed, minter.TotalMinted)
	}

	return nil
}

func calcMintedTokens(m Minter) sdk.Uint {
	fixedMonthsTokens := sdk.NewUint(0)
	calculatedTokensByIntegral := sdk.NewUint(0)
	if m.NormTimePassed.GTE(MonthsInFormula) {
		fixedMonthsTokens.Add((util.ConvertToMicroNolusDec(m.NormTimePassed.Sub(MonthsInFormula))).Mul(FixedMintedAmount))
		calculatedTokensByIntegral.Add(CalcTokensByIntegral(MonthsInFormula).Sub(CalcTokensByIntegral(NormOffset)))
	} else {
		calculatedTokensByIntegral.Add(CalcTokensByIntegral(m.NormTimePassed).Sub(CalcTokensByIntegral(NormOffset)))
	}
	return calculatedTokensByIntegral.Add(fixedMonthsTokens)
}

// Integral:  -1.08319 x^4 + 314.871 x^3 - 44283.6 x^2 + 3.86335×10^6 x
// transformed to: (((-1.08319 x + 314.871) x - 44283.6) x +3.86335×10^6) x
func CalcTokensByIntegral(x sdk.Dec) sdk.Uint {
	return util.ConvertToMicroNolusDec(((((QuadCoef.Mul(x).Add(CubeCoef)).Mul(x).Add(SquareCoef)).Mul(x).Add(Coef)).Mul(x)))
}

func GetAbsDiff(a sdk.Uint, b sdk.Uint) sdk.Uint {
	if a.GTE(b) {
		return a.Sub(b)
	}

	return b.Sub(a)
}

func DecFromUint(u sdk.Uint) sdk.Dec {
	return sdk.NewDecFromBigInt(u.BigInt())
}
