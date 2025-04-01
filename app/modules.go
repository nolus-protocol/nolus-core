package app

import (
	"cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	sdkparams "github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/ibc-go/modules/capability"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"github.com/Nolus-Protocol/nolus-core/x/mint"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	"github.com/Nolus-Protocol/nolus-core/x/tax"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
	"github.com/Nolus-Protocol/nolus-core/x/vestings"
	vestingstypes "github.com/Nolus-Protocol/nolus-core/x/vestings/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/evm/x/erc20"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/cosmos/evm/x/feemarket"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	"github.com/cosmos/evm/x/vm"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/Nolus-Protocol/nolus-core/x/contractmanager"
	contractmanagermoduletypes "github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
	"github.com/Nolus-Protocol/nolus-core/x/feerefunder"
	feetypes "github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs"
	interchaintxstypes "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
	transferSudo "github.com/Nolus-Protocol/nolus-core/x/transfer"
)

// module account permissions.
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	distrtypes.ModuleName:          nil,
	minttypes.ModuleName:           {authtypes.Minter},
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:            {authtypes.Burner},
	ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	wasmtypes.ModuleName:           {authtypes.Burner},
	vestingstypes.ModuleName:       nil,
	icatypes.ModuleName:            nil,
	feetypes.ModuleName:            nil,

	// Cosmos EVM modules
	evmtypes.ModuleName:       {authtypes.Minter, authtypes.Burner},
	feemarkettypes.ModuleName: nil,
}

// ModuleBasics defines the module BasicManager is in charge of setting up basic,
// non-dependant module elements, such as codec registration
// and genesis verification.
var ModuleBasics = module.NewBasicManager(
	genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distribution.AppModuleBasic{},
	gov.NewAppModuleBasic(
		[]govclient.ProposalHandler{
			paramsclient.ProposalHandler,
		},
	),
	sdkparams.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
	feegrantmodule.AppModuleBasic{},
	ibc.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	evidence.AppModuleBasic{},
	transferSudo.AppModuleBasic{},
	vesting.AppModuleBasic{},
	wasm.AppModuleBasic{},
	vestings.AppModuleBasic{},
	tax.AppModuleBasic{},
	ica.AppModuleBasic{},
	interchaintxs.AppModuleBasic{},
	feerefunder.AppModuleBasic{},
	contractmanager.AppModuleBasic{},
	authzmodule.AppModuleBasic{},
	consensus.AppModuleBasic{},
	ibctm.AppModuleBasic{},
	vm.AppModuleBasic{},
	feemarket.AppModuleBasic{},
	erc20.AppModuleBasic{},
)

func appModules(
	app *App,
	encodingConfig EncodingConfig,
	skipGenesisInvariants bool,
) []module.AppModule {
	appCodec := encodingConfig.Marshaler

	return []module.AppModule{
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app,
			encodingConfig.TxConfig,
		),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, encodingConfig.InterfaceRegistry),
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(*app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, *app.BankKeeper, *app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, *app.FeegrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		distribution.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		evidence.NewAppModule(*app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		sdkparams.NewAppModule(*app.ParamsKeeper),
		tax.NewAppModule(appCodec, *app.TaxKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(taxmoduletypes.ModuleName)),
		app.AppKeepers.TransferModule,
		app.AppKeepers.VestingsModule,
		app.AppKeepers.IcaModule,
		app.AppKeepers.InterchainTxsModule,
		app.AppKeepers.FeeRefunderModule,
		app.AppKeepers.ContractManagerModule,
		consensus.NewAppModule(appCodec, *app.AppKeepers.ConsensusParamsKeeper),
		vm.NewAppModule(app.EVMKeeper, app.AppKeepers.AccountKeeper, app.GetSubspace(evmtypes.ModuleName)),
		feemarket.NewAppModule(app.AppKeepers.FeeMarketKeeper, app.GetSubspace(feemarkettypes.ModuleName)),
		erc20.NewAppModule(app.AppKeepers.Erc20Keeper, *app.AppKeepers.AccountKeeper, app.GetSubspace(erc20types.ModuleName)),
	}
}

// simulationModules returns modules for simulation manager
// define the order of the modules for deterministic simulations.
func simulationModules(
	app *App,
	encodingConfig EncodingConfig,
	_ bool,
) []module.AppModuleSimulation {
	appCodec := encodingConfig.Marshaler

	return []module.AppModuleSimulation{
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, encodingConfig.InterfaceRegistry),
		auth.NewAppModule(appCodec, *app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, *app.BankKeeper, *app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, *app.FeegrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, *app.MintKeeper, app.AccountKeeper, app.GetSubspace(minttypes.ModuleName)),
		tax.NewAppModule(appCodec, *app.TaxKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(taxmoduletypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		distribution.NewAppModule(appCodec, *app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		sdkparams.NewAppModule(*app.ParamsKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		ibc.NewAppModule(app.IBCKeeper),
		app.AppKeepers.TransferModule,
		app.AppKeepers.InterchainTxsModule,
	}
}

/*
orderBeginBlockers tells the app's module manager how to set the order of
BeginBlockers, which are run at the beginning of every block.

Interchain Security Requirements:
During begin block slashing happens after distr.BeginBlocker so that
there is nothing left over in the validator fee pool, so as to keep the
CanWithdrawInvariant invariant.
NOTE: staking module is required if HistoricalEntries param > 0
NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
*/

func orderBeginBlockers() []string {
	return []string{
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibcexported.ModuleName,
		genutiltypes.ModuleName,
		banktypes.ModuleName,
		vestingtypes.ModuleName,
		authtypes.ModuleName,
		paramstypes.ModuleName,
		authz.ModuleName,
		ibctransfertypes.ModuleName,
		crisistypes.ModuleName,
		feegrant.ModuleName,
		taxmoduletypes.ModuleName,
		vestingstypes.ModuleName,
		govtypes.ModuleName,
		icatypes.ModuleName,
		interchaintxstypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		wasmtypes.ModuleName,
		feetypes.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		erc20types.ModuleName,
	}
}

func orderEndBlockers() []string {
	return []string{
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		ibcexported.ModuleName,
		paramstypes.ModuleName,
		slashingtypes.ModuleName,
		upgradetypes.ModuleName,
		authtypes.ModuleName,
		capabilitytypes.ModuleName,
		vestingtypes.ModuleName,
		minttypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		authz.ModuleName,
		ibctransfertypes.ModuleName,
		genutiltypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		taxmoduletypes.ModuleName,
		vestingstypes.ModuleName,
		icatypes.ModuleName,
		interchaintxstypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		wasmtypes.ModuleName,
		feetypes.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		erc20types.ModuleName,
	}
}

/*
NOTE: The genutils module must occur after staking so that pools are
properly initialized with tokens from genesis accounts.
NOTE: The genutils module must also occur after auth so that it can access the params from auth.
NOTE: Capability module must occur first so that it can initialize any capabilities
so that other modules that want to create or claim capabilities afterwards in InitChain
can do so safely.
*/
func genesisModuleOrder() []string {
	return []string{
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		taxmoduletypes.ModuleName,
		vestingstypes.ModuleName,
		ibcexported.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		interchaintxstypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		// wasm after ibc transfer
		wasmtypes.ModuleName,
		feetypes.ModuleName,
		consensusparamtypes.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		erc20types.ModuleName,
	}
}
