package app_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/spm/cosmoscmd"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"gitlab-nomo.credissimo.net/nomo/cosmzone/app"
)

var NumSeeds int
var NumTimesToRunPerSeed int

func init() {
	simapp.GetSimulatorFlags()
	flag.IntVar(&NumSeeds, "NumSeeds", 3, "number of random seeds to use")
	flag.IntVar(&NumTimesToRunPerSeed, "NumTimesToRunPerSeed", 5, "number of time to run the simulation per seed")
}

func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestAppStateDeterminism(t *testing.T) {
	if !simapp.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simapp.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = helpers.SimAppChainID

	appHashList := make([]json.RawMessage, NumTimesToRunPerSeed)

	for i := 0; i < NumSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < NumTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simapp.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := tmdb.NewMemDB()
			newApp := app.New(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, simapp.FlagPeriodValue, cosmoscmd.MakeEncodingConfig(app.ModuleBasics), simapp.EmptyAppOptions{}, interBlockCacheOpt())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, NumSeeds, j+1, NumTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				newApp.(*app.App).BaseApp,
				simapp.AppStateFn(newApp.(*app.App).AppCodec(), newApp.(*app.App).SimulationManager()),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simapp.SimulationOperations(newApp.(*app.App), newApp.(*app.App).AppCodec(), config),
				newApp.(*app.App).ModuleAccountAddrs(),
				config,
				newApp.(*app.App).AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simapp.PrintStats(db)
			}

			appHash := newApp.(*app.App).LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, NumSeeds, j+1, NumTimesToRunPerSeed,
				)
			}
		}
	}
}
