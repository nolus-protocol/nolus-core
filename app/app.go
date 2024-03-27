package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/cast"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"

	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"cosmossdk.io/log"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/app/openapiconsole"
	"github.com/Nolus-Protocol/nolus-core/app/params"
	appparams "github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	v053 "github.com/Nolus-Protocol/nolus-core/app/upgrades/v053"
	"github.com/Nolus-Protocol/nolus-core/docs"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	interchaintxstypes "github.com/neutron-org/neutron/v3/x/interchaintxs/types"
)

const (
	Name = "nolus"
)

// DefaultNodeHome default home directories for the application daemon.
var (
	DefaultNodeHome string

	Upgrades = []upgrades.Upgrade{v053.Upgrade}
)

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+appparams.Name)
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp
	keepers.AppKeepers

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry
	encodingConfig    EncodingConfig
	invCheckPeriod    uint

	// the module manager
	mm *module.Manager
	// simulation manager
	sm           *module.SimulationManager
	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app.
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	app := &App{
		BaseApp:           bApp,
		AppKeepers:        keepers.AppKeepers{},
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		encodingConfig:    EncodingConfig(encodingConfig),
	}

	app.NewAppKeepers(
		appCodec,
		bApp,
		encodingConfig.Amino,
		encodingConfig.InterfaceRegistry,
		maccPerms,
		app.BlockedAddrs(),
		skipUpgradeHeights,
		homePath,
		invCheckPeriod,
		appOpts,
		params.Bech32PrefixAccAddr,
	)

	// TODO: decide if we want textual sign mode (https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md#textual-sign-mode)
	// enabledSignModes := append(tx.DefaultSignModes, sigtypes.SignMode_SIGN_MODE_TEXTUAL)
	// txConfigOpts := tx.ConfigOptions{
	// 	EnabledSignModes:           enabledSignModes,
	// 	TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper),
	// }
	// txConfig, err := tx.NewTxConfigWithOptions(
	// 	appCodec,
	// 	txConfigOpts,
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to create new TxConfig with options: %v", err)
	// }
	// app.txConfig = txConfig

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(appModules(app, encodingConfig, skipGenesisInvariants)...)

	//TODO: decide if we need this
	// app.mm.NewBasicManagerFromManager

	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	// Tell the app's module manager how to set the order of BeginBlockers, which are run at the beginning of every block.
	app.mm.SetOrderBeginBlockers(orderBeginBlockers()...)

	app.SetPreBlocker(app.PreBlocker)

	app.mm.SetOrderEndBlockers(orderEndBlockers()...)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(orderInitBlockers()...)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md#app-wiring
	// For app.go without dependency injection(valid for nolus), add the following lines to your app.go in order to provide newer gRPC services:
	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	app.sm = module.NewSimulationManager(simulationModules(app, encodingConfig, skipGenesisInvariants)...)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
				TxFeeChecker:    app.TaxKeeper.CustomTxFeeChecker, // when nil is provided NewDeductFeeDecorator uses default checkTxFeeWithValidatorMinGasPrices
				FeegrantKeeper:  app.FeegrantKeeper,
			},
			BankKeeper:        app.BankKeeper,
			TaxKeeper:         *app.TaxKeeper,
			TxCounterStoreKey: app.GetKVStoreKey()[wasmtypes.StoreKey],
			WasmConfig:        &app.WasmConfig,
			IBCKeeper:         app.IBCKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetInitChainer(app.InitChainer)

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.mm` and `app.configurator` are set.
	app.setupUpgradeHandlers()
	app.setupUpgradeStoreLoaders()

	// must be before Loading version
	// requires the snapshot store to be created and registered as a BaseAppOption
	// see cmd/wasmd/root.go: 206 - 214 approx
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})
		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			panic(err)
		}
	}

	// for local instances - set storage param for the interchain txs module - IcaRegistrationFeeFirstCodeID
	app.SetInterchainTxsLocalChain()

	return app
}

func (app *App) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx, req)
}

func (app *App) SetInterchainTxsLocalChain() {
	// ChainID gets chainID from private fields of BaseApp
	chainID := reflect.ValueOf(app.BaseApp).Elem().FieldByName("chainID").String()

	// If chain is not at block 0 or it's not a local chain, we don't set the storage param
	if app.LastBlockHeight() != 0 || !strings.Contains(chainID, "local") {
		return
	}
	store := app.CommitMultiStore()
	storeInterchaintxs := store.GetKVStore(app.AppKeepers.GetKey(interchaintxstypes.StoreKey))
	// set an extremely high number for the first code id that will be taxed with fee for opening an ICA
	// these bytes are equal to 1684300900 when parsed with sdk.BigEndianToUint64. It's a number that is unlikely to be reached as a code id
	// if we decide to charge a fee for opening an ICA, we can set this to a lower number in the future
	bytesIcaRegistrationFirstCode := []byte{0, 0, 0, 0, 100, 100, 100, 100}
	storeInterchaintxs.Set(interchaintxstypes.ICARegistrationFeeFirstCodeID, bytesIcaRegistrationFirstCode)
}

func (app *App) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				&app.AppKeepers,
				app.appCodec,
			),
		)
	}
}

func (app *App) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
		}
	}
}

// Name returns the name of the App.
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application.
func (app *App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// BeginBlocker application updates every begin block.
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block.
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// InitChainer application update at chain initialization.
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}
	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap()); err != nil {
		return nil, err
	}
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height.
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

//TODO: autocli
// AutoCliOpts returns the autocli options for the app.
// func (app *App) AutoCliOpts() autocli.AppOptions {
// 	modules := make(map[string]appmodule.AppModule, 0)
// 	for _, m := range app.mm.Modules {
// 		if moduleWithName, ok := m.(module.HasName); ok {
// 			moduleName := moduleWithName.Name()
// 			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
// 				modules[moduleName] = appModule
// 			}
// 		}
// 	}

// 	return autocli.AppOptions{
// 		Modules:               modules,
// 		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
// 		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
// 		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
// 		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
// 	}
// }

// AppCodec returns Gaia's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Gaia's InterfaceRegistry.
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface.
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register nodeservice grpc-gateway routes.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/openapi.yml", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapiconsole.Handler(Name, "/static/openapi.yml"))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// GetMaccPerms returns a copy of the module account permissions.
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}
