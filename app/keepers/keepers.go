package keepers

import (
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamstypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/Nolus-Protocol/nolus-core/wasmbinding"
	mintkeeper "github.com/Nolus-Protocol/nolus-core/x/mint/keeper"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxmodulekeeper "github.com/Nolus-Protocol/nolus-core/x/tax/keeper"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	"github.com/Nolus-Protocol/nolus-core/x/vestings"
	vestingskeeper "github.com/Nolus-Protocol/nolus-core/x/vestings/keeper"
	vestingstypes "github.com/Nolus-Protocol/nolus-core/x/vestings/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/neutron-org/neutron/x/contractmanager"
	contractmanagermodulekeeper "github.com/neutron-org/neutron/x/contractmanager/keeper"
	contractmanagermoduletypes "github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/neutron-org/neutron/x/feerefunder"
	feerefunderkeeper "github.com/neutron-org/neutron/x/feerefunder/keeper"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchainqueries"
	interchainquerieskeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	interchainqueriestypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	"github.com/neutron-org/neutron/x/interchaintxs"
	interchaintxskeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	transferSudo "github.com/neutron-org/neutron/x/transfer"
	wrapkeeper "github.com/neutron-org/neutron/x/transfer/keeper"
)

var (
	// WasmProposalsEnabled enables all x/wasm proposals when it's value is "true"
	// and EnableSpecificWasmProposals is empty. Otherwise, all x/wasm proposals
	// are disabled.
	WasmProposalsEnabled = "true"

	// EnableSpecificWasmProposals, if set, must be comma-separated list of values
	// that are all a subset of "EnableAllProposals", which takes precedence over
	// WasmProposalsEnabled.
	//
	// See: https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificWasmProposals = ""

	// EmptyWasmOpts defines a type alias for a list of wasm options.
	EmptyWasmOpts []wasm.Option
)

var (
	// If EnabledSpecificProposals is "", and this is "true", then enable all x/wasm proposals.
	// If EnabledSpecificProposals is "", and this is not "true", then disable all x/wasm proposals.
	ProposalsEnabled = "true"
	// If set to non-empty string it must be comma-separated list of values that are all a subset
	// of "EnableAllProposals" (takes precedence over ProposalsEnabled)
	// https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificProposals = ""
)

// GetWasmEnabledProposals parses the WasmProposalsEnabled and
// EnableSpecificWasmProposals values to produce a list of enabled proposals to
// pass into the application.
func GetWasmEnabledProposals() []wasm.ProposalType {
	if EnableSpecificWasmProposals == "" {
		if WasmProposalsEnabled == "true" {
			return wasm.EnableAllProposals
		}

		return wasm.DisableAllProposals
	}

	chunks := strings.Split(EnableSpecificWasmProposals, ",")

	proposals, err := wasm.ConvertToProposals(chunks)
	if err != nil {
		panic(err)
	}

	return proposals
}

type AppKeepers struct {
	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         *authkeeper.AccountKeeper
	BankKeeper            *bankkeeper.BaseKeeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        *slashingkeeper.Keeper
	DistrKeeper           *distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          *paramskeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	ICAControllerKeeper   *icacontrollerkeeper.Keeper
	EvidenceKeeper        *evidencekeeper.Keeper
	TransferKeeper        *wrapkeeper.KeeperTransferWrapper
	FeeRefunderKeeper     *feerefunderkeeper.Keeper
	ConsensusParamsKeeper *consensusparamskeeper.Keeper
	AuthzKeeper           *authzkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedInterchainTxsKeeper capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper          capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper

	MintKeeper     *mintkeeper.Keeper
	TaxKeeper      *taxmodulekeeper.Keeper
	VestingsKeeper *vestingskeeper.Keeper

	InterchainTxsKeeper     *interchaintxskeeper.Keeper
	InterchainQueriesKeeper *interchainquerieskeeper.Keeper
	ContractManagerKeeper   *contractmanagermodulekeeper.Keeper

	WasmKeeper wasm.Keeper
	WasmConfig wasmtypes.WasmConfig

	// Modules
	ContractManagerModule   contractmanager.AppModule
	InterchainTxsModule     interchaintxs.AppModule
	InterchainQueriesModule interchainqueries.AppModule
	TransferModule          transferSudo.AppModule
	FeeRefunderModule       feerefunder.AppModule
	VestingsModule          vestings.AppModule
	IcaModule               ica.AppModule
	AuthzModule             authzmodule.AppModule
}

func (appKeepers *AppKeepers) NewAppKeepers(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	cdc *codec.LegacyAmino,
	interfaceRegistry codectypes.InterfaceRegistry,
	maccPerms map[string][]string,
	blockedAddress map[string]bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	appOpts servertypes.AppOptions,
	bech32Prefix string,
) {
	// Set keys KVStoreKey, TransientStoreKey, MemoryStoreKey
	appKeepers.GenerateKeys()

	appKeepers.ParamsKeeper = initParamsKeeper(
		appCodec,
		cdc,
		appKeepers.keys[paramstypes.StoreKey],
		appKeepers.tkeys[paramstypes.TStoreKey],
	)

	consensusKeeper := consensusparamskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[consensusparamstypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.ConsensusParamsKeeper = &consensusKeeper

	bApp.SetParamStore(appKeepers.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec, appKeepers.keys[capabilitytypes.StoreKey], appKeepers.memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	appKeepers.ScopedIBCKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	appKeepers.ScopedTransferKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	appKeepers.ScopedInterchainTxsKeeper = appKeepers.CapabilityKeeper.ScopeToModule(interchaintxstypes.ModuleName)
	appKeepers.ScopedICAControllerKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	appKeepers.ScopedWasmKeeper = appKeepers.CapabilityKeeper.ScopeToModule(wasm.ModuleName)

	// seal capabilities after scoping modules
	appKeepers.CapabilityKeeper.Seal()

	appKeepers.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[crisistypes.StoreKey],
		invCheckPeriod,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Add normal keepers
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		appKeepers.keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		bech32Prefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.AccountKeeper = &accountKeeper

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		appKeepers.keys[banktypes.StoreKey],
		appKeepers.AccountKeeper,
		blockedAddress,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.BankKeeper = &bankKeeper

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[stakingtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.StakingKeeper = stakingKeeper

	mintKeeper := mintkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[minttypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.MintKeeper = &mintKeeper

	distrKeeper := distrkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[distrtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.DistrKeeper = &distrKeeper

	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec,
		cdc,
		appKeepers.keys[slashingtypes.StoreKey],
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.SlashingKeeper = &slashingKeeper

	// register the staking hooks
	appKeepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			appKeepers.DistrKeeper.Hooks(),
			appKeepers.SlashingKeeper.Hooks()),
	)

	// UpgradeKeeper must be created before IBCKeeper
	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		appKeepers.keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// UpgradeKeeper must be created before IBCKeeper
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibcexported.StoreKey],
		appKeepers.GetSubspace(ibcexported.ModuleName),
		appKeepers.StakingKeeper,
		appKeepers.UpgradeKeeper,
		appKeepers.ScopedIBCKeeper,
	)

	appKeepers.ContractManagerKeeper = contractmanagermodulekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[contractmanagermoduletypes.StoreKey],
		appKeepers.keys[contractmanagermoduletypes.MemStoreKey],
		appKeepers.WasmKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.ContractManagerModule = contractmanager.NewAppModule(appCodec, *appKeepers.ContractManagerKeeper)

	appKeepers.FeeRefunderKeeper = feerefunderkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[feetypes.StoreKey],
		appKeepers.memKeys[feetypes.MemStoreKey],
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.FeeRefunderModule = feerefunder.NewAppModule(appCodec, *appKeepers.FeeRefunderKeeper, appKeepers.AccountKeeper, appKeepers.BankKeeper)

	transferKeeper := wrapkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.ScopedTransferKeeper,
		appKeepers.FeeRefunderKeeper,
		contractmanager.NewSudoLimitWrapper(appKeepers.ContractManagerKeeper, &appKeepers.WasmKeeper),
	)
	appKeepers.TransferKeeper = &transferKeeper
	appKeepers.TransferModule = transferSudo.NewAppModule(transferKeeper)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	appKeepers.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[evidencetypes.StoreKey],
		appKeepers.StakingKeeper,
		appKeepers.SlashingKeeper,
	)

	icaControllerKeeper := icacontrollerkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[icacontrollertypes.StoreKey],
		appKeepers.GetSubspace(icacontrollertypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedICAControllerKeeper,
		bApp.MsgServiceRouter(),
	)
	appKeepers.ICAControllerKeeper = &icaControllerKeeper

	appKeepers.IcaModule = ica.NewAppModule(appKeepers.ICAControllerKeeper, nil)

	appKeepers.InterchainQueriesKeeper = interchainquerieskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[interchainqueriestypes.StoreKey],
		appKeepers.keys[interchainqueriestypes.MemStoreKey],
		appKeepers.IBCKeeper,
		appKeepers.BankKeeper,
		appKeepers.ContractManagerKeeper,
		interchainquerieskeeper.Verifier{},
		interchainquerieskeeper.TransactionVerifier{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.InterchainQueriesModule = interchainqueries.NewAppModule(appCodec, *appKeepers.InterchainQueriesKeeper, appKeepers.AccountKeeper, appKeepers.BankKeeper)

	appKeepers.InterchainTxsKeeper = interchaintxskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[interchaintxstypes.StoreKey],
		appKeepers.memKeys[interchaintxstypes.MemStoreKey],
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.ICAControllerKeeper,
		contractmanager.NewSudoLimitWrapper(appKeepers.ContractManagerKeeper, &appKeepers.WasmKeeper),
		appKeepers.FeeRefunderKeeper,
		appKeepers.BankKeeper,
		func(ctx sdk.Context) string { return appKeepers.TaxKeeper.ContractAddress(ctx) },
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.InterchainTxsModule = interchaintxs.NewAppModule(appCodec, *appKeepers.InterchainTxsKeeper, appKeepers.AccountKeeper, appKeepers.BankKeeper)

	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}
	appKeepers.WasmConfig = wasmConfig

	var wasmOpts []wasm.Option
	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := "iterator,staking,stargate,migrate,upgrade,neutron,cosmwasm_1_1,cosmwasm_1_2"
	wasmOpts = append(wasmbinding.RegisterCustomPlugins(appKeepers.InterchainTxsKeeper, appKeepers.InterchainQueriesKeeper, *appKeepers.TransferKeeper, appKeepers.FeeRefunderKeeper, appKeepers.ContractManagerKeeper), wasmOpts...)
	appKeepers.WasmKeeper = wasm.NewKeeper(
		appCodec,
		appKeepers.keys[wasm.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		distrkeeper.NewQuerier(*appKeepers.DistrKeeper),
		appKeepers.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 feerefunder
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedWasmKeeper,
		appKeepers.TransferKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*appKeepers.ParamsKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(appKeepers.IBCKeeper.ClientKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(appKeepers.UpgradeKeeper)).
		AddRoute(ibcexported.RouterKey, ibcclient.NewClientProposalHandler(appKeepers.IBCKeeper.ClientKeeper))

	// The gov proposal types can be individually enabled
	if len(GetWasmEnabledProposals()) != 0 {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(appKeepers.WasmKeeper, GetWasmEnabledProposals()))
	}

	govConfig := govtypes.DefaultConfig()
	// MaxMetadataLen defines the maximum proposal metadata length.
	govConfig.MaxMetadataLen = 20000

	appKeepers.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[govtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		bApp.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	appKeepers.GovKeeper.SetLegacyRouter(govRouter)

	taxKeeper := taxmodulekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[taxmoduletypes.StoreKey],
		appKeepers.keys[taxmoduletypes.MemStoreKey],
		appKeepers.WasmKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.TaxKeeper = &taxKeeper

	appKeepers.VestingsKeeper = vestingskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[vestingstypes.StoreKey],
		appKeepers.keys[vestingstypes.MemStoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.VestingsModule = vestings.NewAppModule(appCodec, *appKeepers.VestingsKeeper)

	transferIBCModule := transferSudo.NewIBCModule(
		*appKeepers.TransferKeeper,
		contractmanager.NewSudoLimitWrapper(appKeepers.ContractManagerKeeper, &appKeepers.WasmKeeper),
	)

	var icaControllerStack ibcporttypes.IBCModule

	icaControllerStack = interchaintxs.NewIBCModule(*appKeepers.InterchainTxsKeeper)
	icaControllerStack = icacontroller.NewIBCMiddleware(icaControllerStack, *appKeepers.ICAControllerKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(ibctransfertypes.ModuleName, transferIBCModule).
		AddRoute(interchaintxstypes.ModuleName, icaControllerStack).
		AddRoute(wasm.ModuleName, wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.ChannelKeeper))
	appKeepers.IBCKeeper.SetRouter(ibcRouter)

	authzKeepper := authzkeeper.NewKeeper(
		appKeepers.keys[authzkeeper.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)
	appKeepers.AuthzKeeper = &authzKeepper
	appKeepers.AuthzModule = authzmodule.NewAppModule(appCodec, authzKeepper, appKeepers.AccountKeeper, appKeepers.BankKeeper, interfaceRegistry)
}

// GetSubspace returns a param subspace for a given module name.
func (appKeepers *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := appKeepers.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// initParamsKeeper init params keeper and its subspaces.
func initParamsKeeper(
	appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey,
) *paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)
	paramsKeeper.Subspace(taxmoduletypes.ModuleName)
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(feetypes.ModuleName)
	paramsKeeper.Subspace(interchaintxstypes.ModuleName)
	paramsKeeper.Subspace(interchainqueriestypes.ModuleName)
	paramsKeeper.Subspace(vestingstypes.ModuleName)

	return &paramsKeeper
}
