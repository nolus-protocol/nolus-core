package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/client/v2/autocli"
	clientv2keyring "cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/CosmWasm/wasmd/x/wasm"

	tmcfg "github.com/cometbft/cometbft/config"
	tmcli "github.com/cometbft/cometbft/libs/cli"

	db "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/Nolus-Protocol/nolus-core/app"
)

type (
	// AppBuilder is a method that allows to build an app.
	AppBuilder func(
		logger log.Logger,
		database db.DB,
		traceStore io.Writer,
		loadLatest bool,
		skipUpgradeHeights map[int64]bool,
		homePath string,
		invCheckPeriod uint,
		encodingConfig app.EncodingConfig,
		appOpts servertypes.AppOptions,
		baseAppOptions ...func(*baseapp.BaseApp),
	) App

	// App represents a Cosmos SDK application that can be run as a server and with an exportable state.
	App interface {
		servertypes.Application
		ExportableApp
	}

	// ExportableApp represents an app with an exportable state.
	ExportableApp interface {
		ExportAppStateAndValidators(
			forZeroHeight bool,
			jailAllowedAddrs []string,
			modulesToExport []string,
		) (servertypes.ExportedApp, error)
		LoadHeight(height int64) error
	}

	// appCreator is an app creator.
	appCreator struct {
		encodingConfig app.EncodingConfig
	}
)

// NewRootCmd creates a new root command for a Cosmos SDK application.
func NewRootCmd(
	appName,
	defaultNodeHome,
	defaultChainID string,
	moduleBasics module.BasicManager,
) (*cobra.Command, app.EncodingConfig) {
	var (
		autoCliOpts        autocli.AppOptions
		moduleBasicManager module.BasicManager
		clientCtx          client.Context
	)
	encodingConfig := app.MakeEncodingConfig(moduleBasics)

	if err := depinject.Inject(
		depinject.Configs(app.AppConfig,
			depinject.Supply(
				log.NewNopLogger(),
				// TODO: should we use simtestutil funciton?
				simtestutil.NewAppOptionsWithFlagHome(tempDir()),
			),
			depinject.Provide(
				ProvideClientContext,
				ProvideKeyring,
			),
		),
		&autoCliOpts,
		&moduleBasicManager,
		&clientCtx,
	); err != nil {
		panic(err)
	}

	rootCmd := &cobra.Command{
		Use:   appName + "d",
		Short: "Nolus",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			clientCtx = clientCtx.WithCmdContext(cmd.Context())
			clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err = config.ReadFromClientConfig(clientCtx)
			if err != nil {
				return err
			}

			// TODO: refer to ProvideClientContext func
			// if !clientCtx.Offline {
			// 	enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
			// 	txConfigOpts := tx.ConfigOptions{
			// 		EnabledSignModes:           enabledSignModes,
			// 		TextualCoinMetadataQueryFn: authtxconfig.NewGRPCCoinMetadataQueryFn(clientCtx),
			// 	}
			// 	txConfig, err := tx.NewTxConfigWithOptions(
			// 		clientCtx.Codec,
			// 		txConfigOpts,
			// 	)
			// 	if err != nil {
			// 		return err
			// 	}

			// 	clientCtx = clientCtx.WithTxConfig(txConfig)
			// }

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, tmcfg.DefaultConfig())
		},
	}

	initRootCmd(
		rootCmd,
		encodingConfig,
	)

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        defaultChainID,
		flags.FlagKeyringBackend: "test",
	})

	return rootCmd, encodingConfig
}

func initRootCmd(
	rootCmd *cobra.Command,
	encodingConfig app.EncodingConfig,
) {
	a := appCreator{encodingConfig}

	gentxModule := app.ModuleBasics[genutiltypes.ModuleName].(genutil.AppModuleBasic)

	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, gentxModule.GenTxValidator, encodingConfig.TxConfig.SigningContext().ValidatorAddressCodec()),
		genutilcli.GenTxCmd(
			app.ModuleBasics,
			encodingConfig.TxConfig,
			banktypes.GenesisBalancesIterator{},
			app.DefaultNodeHome,
			encodingConfig.TxConfig.SigningContext().ValidatorAddressCodec(),
		),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		addGenesisAccountCmd(app.DefaultNodeHome),
		addGenesisWasmMsgCmd(app.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(a.newApp, app.DefaultNodeHome),
		snapshot.Cmd(a.newApp),
	)

	// add server commands
	server.AddCommands(
		rootCmd,
		app.DefaultNodeHome,
		a.newApp,
		a.appExport,
		addModuleInitFlags,
	)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		server.ShowValidatorCmd(),
		server.ShowNodeIDCmd(),
		server.ShowAddressCmd(),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)
}

// queryCommand returns the sub-command to send queries to the app.
func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		server.QueryBlockResultsCmd(),
		server.QueryBlocksCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// txCommand returns the sub-command to send transactions to the app.
func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	wasm.AddModuleInitFlags(startCmd)
}

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			err := f.Value.Set(val)
			if err != nil {
				panic(err)
			}
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}

// newApp creates a new Cosmos SDK app.
func (a appCreator) newApp(
	logger log.Logger,
	database db.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache storetypes.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := db.NewDB("metadata", db.GoLevelDBBackend, snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		logger.Error("application home not set, using DefaultNodeHome")
		homePath = app.DefaultNodeHome
	}

	chainID := cast.ToString(appOpts.Get(flags.FlagChainID))
	if chainID == "" {
		// fallback to genesis chain-id
		appGenesis, err := genutiltypes.AppGenesisFromFile(filepath.Join(homePath, "config", "genesis.json"))
		if err != nil {
			panic(err)
		}

		chainID = appGenesis.ChainID
	}

	return app.New(
		logger,
		database,
		traceStore,
		true,
		appOpts,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshottypes.SnapshotOptions{
			Interval:   cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval)),
			KeepRecent: cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent)),
		}),
		baseapp.SetChainID(chainID),
	)
}

// appExport creates a new simapp (optionally at a given height).
func (a appCreator) appExport(
	logger log.Logger,
	database db.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var exportableApp ExportableApp

	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	exportableApp = app.New(
		logger,
		database,
		traceStore,
		height == -1, // -1: no height provided
		appOpts,
	)

	if height != -1 {
		if err := exportableApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return exportableApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	// The following code snippet is just for reference.

	// WASMConfig defines configuration for the wasm module.
	type WASMConfig struct {
		// This is the maximum sdk gas (wasm and storage) that we allow for any x/wasm "smart" queries
		QueryGasLimit uint64 `mapstructure:"query_gas_limit"`

		// Address defines the gRPC-web server to listen on
		LruSize uint64 `mapstructure:"lru_size"`
	}

	type CustomAppConfig struct {
		serverconfig.Config

		WASM WASMConfig `mapstructure:"wasm"`
	}

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In simapp, we set the min gas prices to 0.
	srvCfg.MinGasPrices = "0.0025unls"

	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
		WASM: WASMConfig{
			LruSize:       1,
			QueryGasLimit: 300000,
		},
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate + `
[wasm]
# This is the maximum sdk gas (wasm and storage) that we allow for any x/wasm "smart" queries
query_gas_limit = 300000
# This is the number of wasm vm instances we keep cached in memory for speed-up
# Warning: this is currently unstable and may lead to crashes, best to keep for 0 unless testing locally
lru_size = 0`

	return customAppTemplate, customAppConfig
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", "nolusd")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	return dir
}

func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfigOpts tx.ConfigOptions,
	legacyAmino *codec.LegacyAmino,
) client.Context {

	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithLegacyAmino(legacyAmino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("") // In simapp, we don't use any prefix for env variables.

	// Read the config again to overwrite the default values with the values from the config file
	clientCtx, _ = config.ReadDefaultValuesFromDefaultClientConfig(clientCtx)

	// textual is enabled by default, we need to re-create the tx config grpc instead of bank keeper.
	txConfigOpts.TextualCoinMetadataQueryFn = authtxconfig.NewGRPCCoinMetadataQueryFn(clientCtx)
	txConfig, err := tx.NewTxConfigWithOptions(clientCtx.Codec, txConfigOpts)
	if err != nil {
		panic(err)
	}
	clientCtx = clientCtx.WithTxConfig(txConfig)

	return clientCtx
}

func ProvideKeyring(clientCtx client.Context, addressCodec address.Codec) (clientv2keyring.Keyring, error) {
	kb, err := client.NewKeyringFromBackend(clientCtx, clientCtx.Keyring.Backend())
	if err != nil {
		return nil, err
	}

	return keyring.NewAutoCLIKeyring(kb)
}
