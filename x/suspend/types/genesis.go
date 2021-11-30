package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

func NewGenesis(feeRate types.Dec, feeCaps types.Coins, feeProceeds types.Coins) *GenesisState {
	return &GenesisState{
		FeeRate:     feeRate,
		FeeCaps:     feeCaps,
		FeeProceeds: feeProceeds,
	}
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		FeeRate:     types.ZeroDec(),
		FeeCaps:     types.Coins{},
		FeeProceeds: types.Coins{},
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.FeeRate.IsNil() || gs.FeeRate.IsNegative() {
		return fmt.Errorf("treasury parameter feeRate must not be nil or negative, received: %s", gs.FeeRate)
	}

	for _, feeCap := range gs.FeeCaps {
		if err := feeCap.Validate(); err != nil {
			return err
		}
	}

	for _, proceeds := range gs.FeeProceeds {
		if err := proceeds.Validate(); err != nil {
			return err
		}
	}
	return nil
}
