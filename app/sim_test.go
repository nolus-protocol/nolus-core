package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"

	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	wasmsim "github.com/CosmWasm/wasmd/x/wasm/simulation"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"

	contractmanagermoduletypes "github.com/neutron-org/neutron/v3/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/v3/x/feerefunder/types"
	interchainqueriestypes "github.com/neutron-org/neutron/v3/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v3/x/interchaintxs/types"
)

const SimAppChainID = "nolus-simapp"

var (
	NumSeeds             int
	NumTimesToRunPerSeed int
)

func init() {
	simcli.GetSimulatorFlags()
	flag.IntVar(&NumSeeds, "NumSeeds", 3, "number of random seeds to use")
	flag.IntVar(&NumTimesToRunPerSeed, "NumTimesToRunPerSeed", 5, "number of time to run the simulation per seed")
}

type StoreKeysPrefixes struct {
	A        storetypes.StoreKey
	B        storetypes.StoreKey
	Prefixes [][]byte
}

func appParamsConfigurationSim(t *testing.T, config *simtypes.Config) {
	pkg, err := build.Default.Import("github.com/CosmWasm/wasmd/x/wasm/keeper", "", build.FindOnly)
	if err != nil {
		t.Fatalf("CosmWasm module path not found: %v", err)
	}

	reflectContractPath := filepath.Join(pkg.Dir, "testdata/reflect_1_1.wasm")
	minDepositBytes, err := json.Marshal(sdk.NewCoins((sdk.NewCoin("unls", sdkmath.NewInt(500000)))))
	if err != nil {
		t.Fatal("Marshaling of sdk coins to be used for min deposit param in gov module failed")
	}

	appParams := simtypes.AppParams{
		wasmsim.OpReflectContractPath: []byte(fmt.Sprintf("\"%s\"", reflectContractPath)),
		govsim.MinDeposit:             minDepositBytes,
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
}

func TestAppStateDeterminism(t *testing.T) {
	sdk.DefaultBondDenom = params.DefaultBondDenom
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = SimAppChainID
	appParamsConfigurationSim(t, &config)

	appHashList := make([]json.RawMessage, NumTimesToRunPerSeed)

	for i := 0; i < NumSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < NumTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simcli.FlagVerboseValue {
				logger = log.NewTestLogger(t)
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			encConfig := MakeEncodingConfig(ModuleBasics)
			newApp := New(
				logger,
				db,
				nil,
				true,
				map[int64]bool{},
				DefaultNodeHome,
				simcli.FlagPeriodValue,
				encConfig,
				simtestutil.EmptyAppOptions{},
				fauxMerkleModeOpt,
				baseapp.SetChainID(SimAppChainID),
				baseapp.SetMinGasPrices("0unls"),
			)

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, NumSeeds, j+1, NumTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				newApp.BaseApp,
				simtestutil.AppStateFn(newApp.AppCodec(), newApp.SimulationManager(), NewDefaultGenesisState(encConfig)),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simtestutil.SimulationOperations(newApp, newApp.AppCodec(), config),
				newApp.BlockedAddrs(),
				config,
				newApp.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simtestutil.PrintStats(db)
			}

			appHash := newApp.LastCommitID().Hash
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

func TestAppImportExport(t *testing.T) {
	sdk.DefaultBondDenom = params.DefaultBondDenom
	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID
	appParamsConfigurationSim(t, &config)

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	encConf := MakeEncodingConfig(ModuleBasics)
	nolusApp := New(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		dir,
		simcli.FlagPeriodValue,
		encConf,
		simtestutil.EmptyAppOptions{},
		fauxMerkleModeOpt,
		baseapp.SetChainID(SimAppChainID),
	)
	require.Equal(t, Name, nolusApp.Name())

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		nolusApp.BaseApp,
		simtestutil.AppStateFn(nolusApp.AppCodec(), nolusApp.SimulationManager(), NewDefaultGenesisState(encConf)),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(nolusApp, nolusApp.AppCodec(), config),
		nolusApp.ModuleAccountAddrs(),
		config,
		nolusApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(nolusApp, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	t.Log("exporting genesis...")

	exported, err := nolusApp.ExportAppStateAndValidators(false, []string{}, nolusApp.mm.ModuleNames())
	require.NoError(t, err)

	t.Log("importing genesis...")

	newDB, newDir, _, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim-2", "Simulation-2", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()
	newNolusApp := New(
		log.NewNopLogger(),
		newDB,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		simcli.FlagPeriodValue,
		MakeEncodingConfig(ModuleBasics),
		simtestutil.EmptyAppOptions{},
		fauxMerkleModeOpt,
		baseapp.SetChainID(SimAppChainID),
	)
	require.Equal(t, Name, newNolusApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := nolusApp.NewContextLegacy(true, tmproto.Header{Height: nolusApp.LastBlockHeight()})
	ctxB := newNolusApp.NewContextLegacy(true, tmproto.Header{Height: nolusApp.LastBlockHeight()})
	_, err = newNolusApp.mm.InitGenesis(ctxB, nolusApp.AppCodec(), genesisState)

	if err != nil {
		if strings.Contains(err.Error(), "validator set is empty after InitGenesis") {
			logger.Info("Skipping simulation as all validators have been unbonded")
			logger.Info("err", err, "stacktrace", string(debug.Stack()))
			return
		}
	}

	err = newNolusApp.StoreConsensusParams(ctxB, exported.ConsensusParams)
	if err != nil {
		logger.Info("err on store consensus params: ", err)
		return
	}

	t.Log("comparing stores...")

	keys := nolusApp.AppKeepers.GetKVStoreKeys()
	newKeys := newNolusApp.AppKeepers.GetKVStoreKeys()
	storeKeysPrefixes := []StoreKeysPrefixes{
		{keys[authtypes.StoreKey], newKeys[authtypes.StoreKey], [][]byte{}},
		{
			keys[stakingtypes.StoreKey], newKeys[stakingtypes.StoreKey],
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey, stakingtypes.UnbondingIDKey, stakingtypes.UnbondingIndexKey, stakingtypes.UnbondingTypeKey, stakingtypes.ValidatorUpdatesKey,
			},
		},
		{keys[slashingtypes.StoreKey], newKeys[slashingtypes.StoreKey], [][]byte{}},
		{keys[minttypes.StoreKey], newKeys[minttypes.StoreKey], [][]byte{}},
		{keys[distrtypes.StoreKey], newKeys[distrtypes.StoreKey], [][]byte{}},
		{keys[banktypes.StoreKey], newKeys[banktypes.StoreKey], [][]byte{banktypes.BalancesPrefix}},
		{keys[paramstypes.StoreKey], newKeys[paramstypes.StoreKey], [][]byte{}},
		{keys[govtypes.StoreKey], newKeys[govtypes.StoreKey], [][]byte{}},
		{keys[evidencetypes.StoreKey], newKeys[evidencetypes.StoreKey], [][]byte{}},
		{keys[capabilitytypes.StoreKey], newKeys[capabilitytypes.StoreKey], [][]byte{}},
		{keys[ibcexported.StoreKey], newKeys[ibcexported.StoreKey], [][]byte{}},
		{keys[ibctransfertypes.StoreKey], newKeys[ibctransfertypes.StoreKey], [][]byte{}},
		{keys[feetypes.StoreKey], newKeys[feetypes.StoreKey], [][]byte{}},
		{keys[minttypes.StoreKey], newKeys[minttypes.StoreKey], [][]byte{}},
		{keys[taxmoduletypes.StoreKey], newKeys[taxmoduletypes.StoreKey], [][]byte{}},
		{keys[interchaintxstypes.StoreKey], newKeys[interchaintxstypes.StoreKey], [][]byte{}},
		{keys[contractmanagermoduletypes.StoreKey], newKeys[contractmanagermoduletypes.StoreKey], [][]byte{}},
		{keys[interchainqueriestypes.StoreKey], newKeys[interchainqueriestypes.StoreKey], [][]byte{}},
		{keys[icacontrollertypes.StoreKey], newKeys[icacontrollertypes.StoreKey], [][]byte{}},
		{keys[wasmtypes.StoreKey], newKeys[wasmtypes.StoreKey], [][]byte{wasmtypes.TXCounterPrefix}},
	}

	// diff both stores
	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := simtestutil.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		t.Logf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Len(t, failedKVAs, 0, GetSimulationLog(skp.A.Name(), nolusApp.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr simtypes.StoreDecoderRegistry, kvAs, kvBs []kv.Pair) (log string) {
	for i := 0; i < len(kvAs); i++ {
		if len(kvAs[i].Value) == 0 && len(kvBs[i].Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		decoder, ok := sdr[storeName]
		if ok {
			log += decoder(kvAs[i], kvBs[i])
		} else {
			log += fmt.Sprintf("store A %q => %q\nstore B %q => %q\n", kvAs[i].Key, kvAs[i].Value, kvBs[i].Key, kvBs[i].Value)
		}
	}

	return log
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}
