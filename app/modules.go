package app

import (
	"github.com/tendermint/spm/cosmoscmd"

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
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	appparams "github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/x/mint"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	"github.com/Nolus-Protocol/nolus-core/x/tax"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	"github.com/Nolus-Protocol/nolus-core/x/vestings"
	vestingstypes "github.com/Nolus-Protocol/nolus-core/x/vestings/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"

	"github.com/neutron-org/neutron/x/contractmanager"
	contractmanagermoduletypes "github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/neutron-org/neutron/x/feerefunder"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchainqueries"
	interchainqueriestypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	"github.com/neutron-org/neutron/x/interchaintxs"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	transferSudo "github.com/neutron-org/neutron/x/transfer"
)

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler

	govProposalHandlers = append(
		govProposalHandlers,
		paramsclient.ProposalHandler,
		distrclient.ProposalHandler,
		upgradeclient.ProposalHandler,
		upgradeclient.CancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
	)

	govProposalHandlers = append(govProposalHandlers, wasmclient.ProposalHandlers...)

	return govProposalHandlers
}

// module account permissions.
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:        nil,
	distrtypes.ModuleName:             nil,
	minttypes.ModuleName:              {authtypes.Minter},
	stakingtypes.BondedPoolName:       {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:    {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:               {authtypes.Burner},
	ibctransfertypes.ModuleName:       {authtypes.Minter, authtypes.Burner},
	wasm.ModuleName:                   {authtypes.Burner},
	vestingstypes.ModuleName:          nil,
	icatypes.ModuleName:               nil,
	interchainqueriestypes.ModuleName: nil,
	feetypes.ModuleName:               nil,
}

// ModuleBasics defines the module BasicManager is in charge of setting up basic,
// non-dependant module elements, such as codec registration
// and genesis verification.
var ModuleBasics = module.NewBasicManager(
	genutil.AppModuleBasic{},
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distribution.AppModuleBasic{},
	gov.NewAppModuleBasic(getGovProposalHandlers()...),
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
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
	interchainqueries.AppModuleBasic{},
	feerefunder.AppModuleBasic{},
	contractmanager.AppModuleBasic{},
	authzmodule.AppModuleBasic{},
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
			app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, encodingConfig.InterfaceRegistry),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		distribution.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		wasm.NewAppModule(
			appCodec, &app.WasmKeeper,
			app.AccountKeeper, app.BankKeeper,
		),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		app.AppKeepers.TransferModule,
		app.AppKeepers.VestingsModule,
		app.AppKeepers.TaxModule,
		app.AppKeepers.IcaModule,
		app.AppKeepers.InterchainQueriesModule,
		app.AppKeepers.InterchainTxsModule,
		app.AppKeepers.FeeRefunderModule,
		app.AppKeepers.ContractManagerModule,
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
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, encodingConfig.InterfaceRegistry),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
		tax.NewAppModule(appCodec, app.TaxKeeper, app.AccountKeeper, app.BankKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		distribution.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		params.NewAppModule(app.ParamsKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		wasm.NewAppModule(
			appCodec, &app.WasmKeeper,
			app.AccountKeeper, app.BankKeeper,
		),
		ibc.NewAppModule(app.IBCKeeper),
		app.AppKeepers.TransferModule,
		app.AppKeepers.InterchainQueriesModule,
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
		// upgrades should be run first
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibchost.ModuleName,
		genutiltypes.ModuleName,
		banktypes.ModuleName,
		vestingtypes.ModuleName,
		authtypes.ModuleName,
		paramstypes.ModuleName,
		authz.ModuleName,
		ibctransfertypes.ModuleName,
		crisistypes.ModuleName,
		taxmoduletypes.ModuleName,
		vestingstypes.ModuleName,
		govtypes.ModuleName,
		icatypes.ModuleName,
		interchaintxstypes.ModuleName,
		interchainqueriestypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		wasm.ModuleName,
		feetypes.ModuleName,
	}
}

func orderEndBlockers() []string {
	return []string{
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		ibchost.ModuleName,
		paramstypes.ModuleName,
		slashingtypes.ModuleName,
		upgradetypes.ModuleName,
		authtypes.ModuleName,
		capabilitytypes.ModuleName,
		vestingtypes.ModuleName,
		minttypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		ibctransfertypes.ModuleName,
		genutiltypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		taxmoduletypes.ModuleName,
		vestingstypes.ModuleName,
		icatypes.ModuleName,
		interchaintxstypes.ModuleName,
		interchainqueriestypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		wasm.ModuleName,
		feetypes.ModuleName,
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
func orderInitBlockers() []string {
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
		ibchost.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		interchainqueriestypes.ModuleName,
		interchaintxstypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		// wasm after ibc transfer
		wasm.ModuleName,
		feetypes.ModuleName,
	}
}
