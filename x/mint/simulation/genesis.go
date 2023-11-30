package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

// GenMaxMintableNanoseconds generates random MaxMintableNanoseconds in range [1-60).
func GenMaxMintableNanoseconds(r *rand.Rand) sdkmath.Uint {
	return sdkmath.NewUint(uint64(time.Second.Nanoseconds() * int64(r.Intn(59)+1)))
}

// RandomizedGenState generates a random GenesisState for mint.
func RandomizedGenState(simState *module.SimulationState) {
	// minter
	var maxMintableNSecs sdkmath.Uint
	simState.AppParams.GetOrGenerate(
		simState.Cdc, string(types.KeyMaxMintableNanoseconds), &maxMintableNSecs, simState.Rand,
		func(r *rand.Rand) { maxMintableNSecs = GenMaxMintableNanoseconds(r) },
	)
	mintDenom := params.DefaultBondDenom
	params := types.NewParams(mintDenom, maxMintableNSecs)

	mintGenesis := types.NewGenesisState(types.InitialMinter(), params)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
