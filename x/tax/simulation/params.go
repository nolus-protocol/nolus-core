package simulation

import (
	"fmt"
	"math/rand"

	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyFeeRate),
			func(r *rand.Rand) string {
				return fmt.Sprint(GenRandomFeeRate(r))
			},
		),
	}
}
