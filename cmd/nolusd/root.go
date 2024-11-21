package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
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

// FlagRejectConfigDefaults defines a flag to reject some select defaults that override what is in the config file.
const FlagRejectConfigDefaults = "reject-config-defaults"

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

	// SectionKeyValue is used for modifying node config with recommended values.
	SectionKeyValue struct {
		Section string
		Key     string
		Value   any
	}
)

var (
	recommendedAppTomlValues = []SectionKeyValue{
		{
			Section: "wasm",
			Key:     "query_gas_limit",
			Value:   "5000000",
		},
	}

	recommendedConfigTomlValues = []SectionKeyValue{
		{
			Section: "p2p",
			Key:     "flush_throttle_timeout",
			Value:   "80ms",
		},
		{
			Section: "consensus",
			Key:     "timeout_commit",
			Value:   "3s",
		},
		{
			Section: "consensus",
			Key:     "timeout_propose",
			Value:   "3s",
		},
		{
			Section: "consensus",
			Key:     "peer_gossip_sleep_duration",
			Value:   "50ms",
		},
	}
)

// NewRootCmd creates a new root command for a Cosmos SDK application.
func NewRootCmd(
	appName,
	defaultNodeHome,
	defaultChainID string,
	moduleBasics module.BasicManager,
) (*cobra.Command, app.EncodingConfig) {
	encodingConfig := app.MakeEncodingConfig(moduleBasics)

	// create a temporary application for use in constructing query + tx commands
	tempDir := tempDir()

	tempApp := app.New(
		log.NewNopLogger(),
		db.NewMemDB(),
		nil,
		true,
		nil,
		tempDir,
		0,
		encodingConfig,
		simtestutil.NewAppOptionsWithFlagHome(tempDir),
	)
	defer func() {
		if err := tempApp.Close(); err != nil {
			panic(err)
		}
	}()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:   appName + "d",
		Short: "Nolus",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx = initClientCtx.WithCmdContext(cmd.Context())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if !initClientCtx.Offline {
				enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
				txConfigOpts := tx.ConfigOptions{
					EnabledSignModes:           enabledSignModes,
					TextualCoinMetadataQueryFn: authtxconfig.NewGRPCCoinMetadataQueryFn(initClientCtx),
				}
				txConfig, err := tx.NewTxConfigWithOptions(
					initClientCtx.Codec,
					txConfigOpts,
				)
				if err != nil {
					return err
				}

				initClientCtx = initClientCtx.WithTxConfig(txConfig)
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
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

	autoCliOpts := tempApp.AutoCliOpts()
	initClientCtx, err := config.ReadDefaultValuesFromDefaultClientConfig(initClientCtx)
	if err != nil {
		panic(err)
	}
	autoCliOpts.Keyring, err = keyring.NewAutoCLIKeyring(initClientCtx.Keyring)
	if err != nil {
		panic(err)
	}
	autoCliOpts.ClientCtx = initClientCtx

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

	for i, cmd := range rootCmd.Commands() {
		if cmd.Name() == "start" {
			startRunE := cmd.RunE

			// Instrument start command pre run hook with custom logic
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				serverCtx := server.GetServerContextFromCmd(cmd)

				// Get flag value for rejecting config defaults
				rejectConfigDefaults := serverCtx.Viper.GetBool(FlagRejectConfigDefaults)

				// overwrite config.toml and app.toml values, if rejectConfigDefaults is false
				if !rejectConfigDefaults {
					// Add ctx logger line to indicate that config.toml and app.toml values are being overwritten
					serverCtx.Logger.Info("Overwriting config.toml and app.toml values with some recommended defaults. To prevent this, set the --reject-config-defaults flag to true.")

					err := overwriteConfigTomlValues(serverCtx)
					if err != nil {
						return err
					}

					err = overwriteAppTomlValues(serverCtx)
					if err != nil {
						return err
					}
				}

				return startRunE(cmd, args)
			}

			rootCmd.Commands()[i] = cmd
			break
		}
	}
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
	startCmd.Flags().Bool(FlagRejectConfigDefaults, false, "Reject some select recommended default values from being automatically set in the config.toml and app.toml")
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
		skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		a.encodingConfig,
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
		map[int64]bool{},
		homePath,
		uint(1),
		a.encodingConfig,
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

// overwriteConfigTomlValues overwrites config.toml values. Returns error if config.toml does not exist
//
// Currently, overwrites:
// - timeout_commit
//
// Also overwrites the respective viper config value.
//
// Silently handles and skips any error/panic due to write permission issues.
// No-op otherwise.
func overwriteConfigTomlValues(serverCtx *server.Context) error {
	// Get paths to config.toml and config parent directory
	rootDir := serverCtx.Viper.GetString(tmcli.HomeFlag)

	configParentDirPath := filepath.Join(rootDir, "config")
	configFilePath := filepath.Join(configParentDirPath, "config.toml")

	fileInfo, err := os.Stat(configFilePath)
	if err != nil {
		// something besides a does not exist error
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read in %s: %w", configFilePath, err)
		}
	} else {
		// config.toml exists

		// Check if each key is already set to the recommended value
		// If it is, we don't need to overwrite it and can also skip the app.toml overwrite
		var sectionKeyValuesToWrite []SectionKeyValue

		// Set aside which keys need to be updated in the config.toml
		for _, rec := range recommendedConfigTomlValues {
			currentValue := serverCtx.Viper.Get(rec.Section + "." + rec.Key)
			if currentValue != rec.Value {
				// Current value in config.toml is not the recommended value
				// Set the value in viper to the recommended value
				// and add it to the list of key values we will overwrite in the config.toml
				serverCtx.Viper.Set(rec.Section+"."+rec.Key, rec.Value)
				sectionKeyValuesToWrite = append(sectionKeyValuesToWrite, rec)
			}
		}

		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("failed to write to %s: %s\n", configFilePath, err)
			}
		}()

		// Check if the file is writable
		if fileInfo.Mode()&os.FileMode(0o200) != 0 {
			// It will be re-read in server.InterceptConfigsPreRunHandler
			// this may panic for permissions issues. So we catch the panic.
			// Note that this exits with a non-zero exit code if fails to write the file.

			// Write the new config.toml file
			if len(sectionKeyValuesToWrite) > 0 {
				err := OverwriteWithCustomConfig(configFilePath, sectionKeyValuesToWrite)
				if err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("config.toml is not writable. Cannot apply update. Please consider manually changing to the following: %v\n", recommendedConfigTomlValues)
		}
	}
	return nil
}

// overwriteAppTomlValues overwrites app.toml values. Returns error if app.toml does not exist
//
// Currently, overwrites:
// - wasm query_gas_limit
//
// Also overwrites the respective viper config value.
//
// Silently handles and skips any error/panic due to write permission issues.
// No-op otherwise.
func overwriteAppTomlValues(serverCtx *server.Context) error {
	// Get paths to app.toml and config parent directory
	rootDir := serverCtx.Viper.GetString(tmcli.HomeFlag)

	configParentDirPath := filepath.Join(rootDir, "config")
	appFilePath := filepath.Join(configParentDirPath, "app.toml")

	fileInfo, err := os.Stat(appFilePath)
	if err != nil {
		// something besides a does not exist error
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read in %s: %w", appFilePath, err)
		}
	} else {
		// app.toml exists

		// Check if each key is already set to the recommended value
		// If it is, we don't need to overwrite it and can also skip the app.toml overwrite
		var sectionKeyValuesToWrite []SectionKeyValue

		for _, rec := range recommendedAppTomlValues {
			currentValue := serverCtx.Viper.Get(rec.Section + "." + rec.Key)
			if currentValue != rec.Value {
				// Current value in app.toml is not the recommended value
				// Set the value in viper to the recommended value
				// and add it to the list of key values we will overwrite in the app.toml
				serverCtx.Viper.Set(rec.Section+"."+rec.Key, rec.Value)
				sectionKeyValuesToWrite = append(sectionKeyValuesToWrite, rec)
			}
		}

		// Check if the file is writable
		if fileInfo.Mode()&os.FileMode(0o200) != 0 {
			// It will be re-read in server.InterceptConfigsPreRunHandler
			// this may panic for permissions issues. So we catch the panic.
			// Note that this exits with a non-zero exit code if fails to write the file.

			// Write the new app.toml file
			if len(sectionKeyValuesToWrite) > 0 {
				err := OverwriteWithCustomConfig(appFilePath, sectionKeyValuesToWrite)
				if err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("app.toml is not writable. Cannot apply update. Please consider manually changing to the following: %v\n", recommendedAppTomlValues)
		}
	}
	return nil
}

// OverwriteWithCustomConfig searches the respective config file for the given section and key and overwrites the current value with the given value.
func OverwriteWithCustomConfig(configFilePath string, sectionKeyValues []SectionKeyValue) error {
	// Open the file for reading and writing
	file, err := os.OpenFile(configFilePath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a map from the sectionKeyValues array
	// This map will be used to quickly look up the new values for each section and key
	configMap := make(map[string]map[string]string)
	for _, skv := range sectionKeyValues {
		// If the section does not exist in the map, create it
		if _, ok := configMap[skv.Section]; !ok {
			configMap[skv.Section] = make(map[string]string)
		}
		// Add the key and value to the section in the map
		// If the value is a string, add quotes around it
		switch v := skv.Value.(type) {
		case string:
			configMap[skv.Section][skv.Key] = "\"" + v + "\""
		default:
			configMap[skv.Section][skv.Key] = fmt.Sprintf("%v", v)
		}
	}

	// Read the file line by line
	var lines []string
	scanner := bufio.NewScanner(file)
	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		// If the line is a section header, update the current section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
		} else if configMap[currentSection] != nil {
			// If the line is in a section that needs to be overwritten, check each key
			for key, value := range configMap[currentSection] {
				// Split the line into key and value parts
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				// Trim spaces and compare the key part with the target key
				if strings.TrimSpace(parts[0]) == key {
					// If the keys match, overwrite the line with the new key-value pair
					line = key + " = " + value
					break
				}
			}
		}
		// Add the line to the lines slice, whether it was overwritten or not
		lines = append(lines, line)
	}

	// Check for errors from the scanner
	if err := scanner.Err(); err != nil {
		return err
	}

	// Seek to the beginning of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Truncate the file to remove the old content
	err = file.Truncate(0)
	if err != nil {
		return err
	}

	// Write the new lines to the file
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}
