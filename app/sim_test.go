package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/module"
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
	icacontrollertypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v4/modules/core/24-host"

	"github.com/tendermint/spm/cosmoscmd"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmsim "github.com/CosmWasm/wasmd/x/wasm/simulation"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	taxtypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"

	contractmanagermoduletypes "github.com/neutron-org/neutron/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	interchainqueriestypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
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

type StoreKeysPrefixes struct {
	A        sdk.StoreKey
	B        sdk.StoreKey
	Prefixes [][]byte
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

	reflectContractPath := filepath.Join(pkg.Dir, "testdata/reflect_1_1.wasm")
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
			newApp := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, simapp.FlagPeriodValue, cosmoscmd.MakeEncodingConfig(ModuleBasics), simapp.EmptyAppOptions{}, fauxMerkleModeOpt)
			params.SetAddressPrefixes()
			ctx := newApp.(*App).BaseApp.NewUncachedContext(true, tmproto.Header{})
			newApp.(*App).TaxKeeper.SetParams(ctx, taxtypes.DefaultParams())
			newApp.(*App).MintKeeper.SetParams(ctx, minttypes.DefaultParams())
			newApp.(*App).AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
			newApp.(*App).BankKeeper.SetParams(ctx, banktypes.DefaultParams())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, NumSeeds, j+1, NumTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				newApp.(*App).BaseApp,
				simapp.AppStateFn(newApp.(*App).AppCodec(), newApp.(*App).SimulationManager()),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simapp.SimulationOperations(newApp.(*App), newApp.(*App).AppCodec(), config),
				newApp.(*App).BlockedAddrs(),
				config,
				newApp.(*App).AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simapp.PrintStats(db)
			}

			appHash := newApp.(*App).LastCommitID().Hash
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
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	encConf := cosmoscmd.MakeEncodingConfig(ModuleBasics)
	nolusApp := New(logger, db, nil, true, map[int64]bool{}, dir, simapp.FlagPeriodValue, encConf, simapp.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, Name, nolusApp.(*App).Name())

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		nolusApp.(*App).BaseApp,
		AppStateFn(nolusApp.(*App).AppCodec(), nolusApp.(*App).SimulationManager()),
		simtypes.RandomAccounts,
		simapp.SimulationOperations(nolusApp.(*App), nolusApp.(*App).AppCodec(), config),
		nolusApp.(*App).ModuleAccountAddrs(),
		config,
		nolusApp.(*App).AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(nolusApp.(*App), config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}

	t.Log("exporting genesis...")

	exported, err := nolusApp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	t.Log("importing genesis...")

	_, newDB, newDir, _, _, err := simapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()
	newNolusApp := New(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, DefaultNodeHome, simapp.FlagPeriodValue, cosmoscmd.MakeEncodingConfig(ModuleBasics), simapp.EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, Name, newNolusApp.(*App).Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := nolusApp.(*App).NewContext(true, tmproto.Header{Height: nolusApp.(*App).LastBlockHeight()})
	ctxB := newNolusApp.(*App).NewContext(true, tmproto.Header{Height: nolusApp.(*App).LastBlockHeight()})
	newNolusApp.(*App).mm.InitGenesis(ctxB, nolusApp.(*App).AppCodec(), genesisState)
	newNolusApp.(*App).StoreConsensusParams(ctxB, exported.ConsensusParams)

	t.Log("comparing stores...")

	keys := nolusApp.(*App).AppKeepers.GetKVStoreKey()
	newKeys := newNolusApp.(*App).AppKeepers.GetKVStoreKey()
	storeKeysPrefixes := []StoreKeysPrefixes{
		{keys[authtypes.StoreKey], newKeys[authtypes.StoreKey], [][]byte{}},
		{
			keys[stakingtypes.StoreKey], newKeys[stakingtypes.StoreKey],
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey, stakingtypes.UnbondingDelegationKey, stakingtypes.UnbondingDelegationByValIndexKey, stakingtypes.ValidatorsKey,
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
		{keys[ibchost.StoreKey], newKeys[ibchost.StoreKey], [][]byte{}},
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
	normalizeContractInfo(ctxA, nolusApp.(*App))
	normalizeContractInfo(ctxB, newNolusApp.(*App))

	// diff both stores
	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		t.Logf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Len(t, failedKVAs, 0, GetSimulationLog(skp.A.Name(), nolusApp.(*App).SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
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

// AppStateFn returns the initial application state using a genesis or the simulation parameters.
// It panics if the user provides files for both of them.
// If a file is not given for the genesis or the sim params, it creates a randomized one.
func AppStateFn(codec codec.Codec, manager *module.SimulationManager) simtypes.AppStateFn {
	// quick hack to setup app state genesis with our app modules
	simapp.ModuleBasics = ModuleBasics
	if simapp.FlagGenesisTimeValue == 0 { // always set to have a block time
		simapp.FlagGenesisTimeValue = time.Now().Unix()
	}
	return simapp.AppStateFn(codec, manager)
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}
