package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/custom/util"
)

var (
	MintingCap        = util.ConvertToMicroNolusInt(sdk.NewInt(150000000))
	FixedMintedAmount = util.ConvertToMicroNolusInt(sdk.NewInt(103125))
	NormOffset        = sdk.MustNewDecFromStr("0.47")
	MonthsInFormula   = sdk.MustNewDecFromStr("96")
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
			minter.NormTimePassed.String())
	}
	return nil
}
