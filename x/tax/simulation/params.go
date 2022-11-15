package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const (
	keyFeeRate         = "FeeRate"
	keyContractAddress = "ContractAddress"
	keyBaseDenom       = "BaseDenom"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyFeeRate,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenRandomFeeRate(r))
			},
		),
	}
}
