package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"cosmossdk.io/x/simulation"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		simulation.NewSimLegacyParamChange(types.ModuleName, string(types.KeyMaxMintableNanoseconds),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMaxMintableNanoseconds(r).String())
			},
		),
	}
}
