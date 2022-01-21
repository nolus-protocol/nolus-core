package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
	"math/rand"
)

// simulation parameter constants
const (
	BlockHeight = "blockHeight"
	Suspended   = "suspended"
)

// RandomizedGenState generates a random GenesisState for suspend
func RandomizedGenState(simState *module.SimulationState) {
	var blockHeight int64
	var suspended bool
	simState.AppParams.GetOrGenerate(
		simState.Cdc, BlockHeight, &blockHeight, simState.Rand,
		func(r *rand.Rand) { blockHeight = int64(r.Intn(100000)) },
	)
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Suspended, &blockHeight, simState.Rand,
		func(r *rand.Rand) { suspended = r.Intn(2)%2 == 0 },
	)
	tmAddr := ed25519.GenPrivKey().PubKey().Address().String()
	address, _ := sdk.AccAddressFromHex(tmAddr)

	suspendedStateGenesis := types.NewGenesisState(types.NewSuspendedState(address.String(), suspended, blockHeight))

	bz, err := json.MarshalIndent(&suspendedStateGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated suspended parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(suspendedStateGenesis)
}
