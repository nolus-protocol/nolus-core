package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/mint/types"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxMintableNanoseconds),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMaxMintableNanoseconds(r).String())
			},
		),
	}
}
