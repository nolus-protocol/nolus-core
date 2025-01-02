package simulation

import (
	"fmt"
	"math/rand"

	"cosmossdk.io/x/simulation"
	legacytypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		simulation.NewSimLegacyParamChange(types.ModuleName, string(legacytypes.KeyFeeRate),
			func(r *rand.Rand) string {
				return fmt.Sprint(GenRandomFeeRate(r))
			},
		),
	}
}
