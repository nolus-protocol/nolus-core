package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"

	contractmanagermoduletypes "github.com/neutron-org/neutron/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	interchainqueriestypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

// SimAppChainID hardcoded chainID for simulation.
const SimAppChainID = "simulation-app"

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

func TestAppStateDeterminism(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = SimAppChainID

	// pkg, err := build.Default.Import("github.com/CosmWasm/wasmd/x/wasm/keeper", "", build.FindOnly)
	// if err != nil {
	// 	t.Fatalf("CosmWasm module path not found: %v", err)
	// }

	// reflectContractPath := filepath.Join(pkg.Dir, "testdata/reflect_1_1.wasm")
	appParams := simtypes.AppParams{
		// refactor decide how to handle this, problem is importing wasmsim ( maybe upgrade wasmd version)
		// wasmsim.OpReflectContractPath: []byte(fmt.Sprintf("\"%s\"", reflectContractPath)),
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
			if simcli.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := tmdb.NewMemDB()
			newApp := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, simcli.FlagPeriodValue, MakeEncodingConfig(ModuleBasics), simtestutil.EmptyAppOptions{}, fauxMerkleModeOpt)

			// params.SetAddressPrefixes()
			ctx := newApp.NewUncachedContext(true, tmproto.Header{})
			// newApp.TaxKeeper.SetParams(ctx, taxtypes.DefaultParams())
			// newApp.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
			// newApp.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
			// newApp.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, NumSeeds, j+1, NumTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				newApp.BaseApp,
				simtestutil.AppStateFn(newApp.AppCodec(), newApp.SimulationManager(), newApp.mm.ExportGenesis(ctx, newApp.AppCodec())), // refactor: try to find a way to use .DefaultGenesis instead of ExportGenesis
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
	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

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
	nolusApp := New(logger, db, nil, true, map[int64]bool{}, dir, simcli.FlagPeriodValue, encConf, simtestutil.EmptyAppOptions{}, fauxMerkleModeOpt)
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
	newNolusApp := New(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, DefaultNodeHome, simcli.FlagPeriodValue, MakeEncodingConfig(ModuleBasics), simtestutil.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, Name, newNolusApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := nolusApp.NewContext(true, tmproto.Header{Height: nolusApp.LastBlockHeight()})
	ctxB := newNolusApp.NewContext(true, tmproto.Header{Height: nolusApp.LastBlockHeight()})
	newNolusApp.mm.InitGenesis(ctxB, nolusApp.AppCodec(), genesisState)
	newNolusApp.StoreConsensusParams(ctxB, exported.ConsensusParams)

	t.Log("comparing stores...")

	keys := nolusApp.AppKeepers.GetKVStoreKey()
	newKeys := newNolusApp.AppKeepers.GetKVStoreKey()
	storeKeysPrefixes := []StoreKeysPrefixes{
		{keys[authtypes.StoreKey], newKeys[authtypes.StoreKey], [][]byte{}},
		{
			keys[stakingtypes.StoreKey], newKeys[stakingtypes.StoreKey],
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey, stakingtypes.UnbondingDelegationKey, stakingtypes.UnbondingDelegationByValIndexKey, stakingtypes.ValidatorsKey,
				stakingtypes.UnbondingIndexKey, stakingtypes.UnbondingTypeKey, stakingtypes.ValidatorUpdatesKey, stakingtypes.UnbondingIndexKey,
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
		{keys[wasm.StoreKey], newKeys[wasm.StoreKey], [][]byte{}},
	}

	// delete persistent tx counter value
	ctxA.KVStore(keys[wasm.StoreKey]).Delete(wasmtypes.TXCounterPrefix)

	// reset contract code index in source DB for comparison with dest DB
	dropContractHistory := func(s store.KVStore, keys ...[]byte) {
		for _, key := range keys {
			prefixStore := prefix.NewStore(s, key)
			iter := prefixStore.Iterator(nil, nil)
			for ; iter.Valid(); iter.Next() {
				prefixStore.Delete(iter.Key())
			}
			iter.Close()
		}
	}
	prefixes := [][]byte{wasmtypes.ContractCodeHistoryElementPrefix, wasmtypes.ContractByCodeIDAndCreatedSecondaryIndexPrefix}
	dropContractHistory(ctxA.KVStore(keys[wasm.StoreKey]), prefixes...)
	dropContractHistory(ctxB.KVStore(newKeys[wasm.StoreKey]), prefixes...)

	normalizeContractInfo := func(ctx sdk.Context, nApp *App) {
		var index uint64
		nApp.WasmKeeper.IterateContractInfo(ctx, func(address sdk.AccAddress, info wasmtypes.ContractInfo) bool {
			created := &wasmtypes.AbsoluteTxPosition{
				BlockHeight: uint64(0),
				TxIndex:     index,
			}
			info.Created = created
			store := ctx.KVStore(nApp.AppKeepers.GetKVStoreKey()[wasm.StoreKey])
			store.Set(wasmtypes.GetContractAddressKey(address), nApp.appCodec.MustMarshal(&info))
			index++
			return false
		})
	}
	normalizeContractInfo(ctxA, nolusApp)
	normalizeContractInfo(ctxB, newNolusApp)

	// diff both stores
	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		t.Logf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Len(t, failedKVAs, 0, GetSimulationLog(skp.A.Name(), nolusApp.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr sdk.StoreDecoderRegistry, kvAs, kvBs []kv.Pair) (log string) {
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
