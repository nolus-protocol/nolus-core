package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/custom/util"
)

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
	testTotalMinted := CalcIntegral(MonthsInFormula).Sub(CalcIntegral(NormOffset)).Add(TotalMonths.Mul(sdk.NewDecFromInt(FixedMintedAmount)))
	if testTotalMinted.GT(sdk.NewDecFromInt(MintingCap)) {
		return fmt.Errorf("mint parameters for minting formula can not be bigger than MintingCap, %s",
			testTotalMinted)
	}

	return nil
}

// Integral:  -1.08319 x^4 + 314.871 x^3 - 44283.6 x^2 + 3.86335×10^6 x
func CalcIntegral(x sdk.Dec) sdk.Dec {
	return ((((QuadCoef.Mul(x)).Mul(x)).Mul(x)).Add((CubeCoef.Mul(x)).Mul(x).Add(SquareCoef.Mul(x)).Add(Coef)).Mul(x))
}
