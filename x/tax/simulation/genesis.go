package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"
)

// simulation parameter constants
const (
	FeeRate = "FeeRate"
)

// GenRandomFeeRate generates random FeeRate in range [1-100)
func GenRandomFeeRate(r *rand.Rand) int32 {
	return int32(r.Intn(99) + 1)
}

// RandomizedGenState generates a random GenesisState for tax
func RandomizedGenState(simState *module.SimulationState) {
	// tax
	var feeRate int32

	// generate random fee rate between 1 - 100
	simState.AppParams.GetOrGenerate(
		simState.Cdc, FeeRate, &feeRate, simState.Rand,
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
