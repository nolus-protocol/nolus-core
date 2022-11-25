package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"
)

// GenRandomFeeRate generates random FeeRate in range [0-50].
func GenRandomFeeRate(r *rand.Rand) int32 {
	return int32(r.Intn(51))
}

// RandomizedGenState generates a random GenesisState for tax.
func RandomizedGenState(simState *module.SimulationState) {
	var feeRate int32

	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyFeeRate), &feeRate, simState.Rand,
		func(r *rand.Rand) { feeRate = GenRandomFeeRate(r) },
	)
	params := types.NewParams(feeRate, types.DefaultContractAddress, types.DefaultBaseDenom)

	taxGenesis := types.NewGenesisState(params)

	bz, err := json.MarshalIndent(&taxGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated tax parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(taxGenesis)
}
