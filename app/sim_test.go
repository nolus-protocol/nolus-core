package app_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/spm/cosmoscmd"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	wasmsim "github.com/CosmWasm/wasmd/x/wasm/simulation"
	"github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/app/params"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxtypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
)

var (
	NumSeeds             int
	NumTimesToRunPerSeed int
)

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

	pkg, err := build.Default.Import("github.com/CosmWasm/wasmd/x/wasm/keeper", "", build.FindOnly)
	if err != nil {
		t.Fatalf("CosmWasm module path not found: %v", err)
	}

	reflectContractPath := filepath.Join(pkg.Dir, "testdata/reflect.wasm")
	appParams := simtypes.AppParams{
		wasmsim.OpReflectContractPath: []byte(fmt.Sprintf("\"%s\"", reflectContractPath)),
	}
	bz, err := json.Marshal(appParams)
	if err != nil {
		t.Fatal("Marshaling of simulation parameters failed")
	}
	config.ParamsFile = filepath.Join(t.TempDir(), "app-params.json")
	err = os.WriteFile(config.ParamsFile, bz, 0o600)
	if err != nil {
		t.Fatal("Writing of simulation parameters failed")
	}

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
			params.SetAddressPrefixes()
			ctx := newApp.(*app.App).BaseApp.NewUncachedContext(true, tmproto.Header{})
			newApp.(*app.App).TaxKeeper.SetParams(ctx, taxtypes.DefaultParams())
			newApp.(*app.App).MintKeeper.SetParams(ctx, minttypes.DefaultParams())
			newApp.(*app.App).AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
			newApp.(*app.App).BankKeeper.SetParams(ctx, banktypes.DefaultParams())

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
				newApp.(*app.App).BlockedAddrs(),
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
