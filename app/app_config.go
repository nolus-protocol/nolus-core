//go:build !app_v1

package app

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	authzmodulev1 "cosmossdk.io/api/cosmos/authz/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	crisismodulev1 "cosmossdk.io/api/cosmos/crisis/module/v1"
	distrmodulev1 "cosmossdk.io/api/cosmos/distribution/module/v1"
	evidencemodulev1 "cosmossdk.io/api/cosmos/evidence/module/v1"
	feegrantmodulev1 "cosmossdk.io/api/cosmos/feegrant/module/v1"
	genutilmodulev1 "cosmossdk.io/api/cosmos/genutil/module/v1"
	govmodulev1 "cosmossdk.io/api/cosmos/gov/module/v1"
	mintmodulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	paramsmodulev1 "cosmossdk.io/api/cosmos/params/module/v1"
	slashingmodulev1 "cosmossdk.io/api/cosmos/slashing/module/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	upgrademodulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	_ "cosmossdk.io/x/circuit"  // import for side-effects
	_ "cosmossdk.io/x/evidence" // import for side-effects
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	_ "cosmossdk.io/x/feegrant/module" // import for side-effects
	_ "cosmossdk.io/x/nft/module"      // import for side-effects
	_ "cosmossdk.io/x/upgrade"         // import for side-effects
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/module"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side-effects
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	_ "github.com/cosmos/cosmos-sdk/x/authz/module" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"         // import for side-effects
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus" // import for side-effects
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	_ "github.com/cosmos/cosmos-sdk/x/crisis" // import for side-effects
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	_ "github.com/cosmos/cosmos-sdk/x/distribution" // import for side-effects
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	_ "github.com/cosmos/cosmos-sdk/x/group/module" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/params"       // import for side-effects
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	_ "github.com/cosmos/cosmos-sdk/x/slashing" // import for side-effects
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	_ "github.com/cosmos/cosmos-sdk/x/staking" // import for side-effects
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	appparams "github.com/Nolus-Protocol/nolus-core/app/params"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	vestingstypes "github.com/Nolus-Protocol/nolus-core/x/vestings/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	contractmanagermoduletypes "github.com/neutron-org/neutron/v4/x/contractmanager/types"
	feetypes "github.com/neutron-org/neutron/v4/x/feerefunder/types"
	interchainqueriestypes "github.com/neutron-org/neutron/v4/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v4/x/interchaintxs/types"
)

var (
	// module account permissions
	moduleAccPerms = []*authmodulev1.ModuleAccountPermission{
		{Account: authtypes.FeeCollectorName},
		{Account: distrtypes.ModuleName},
		{Account: minttypes.ModuleName, Permissions: []string{authtypes.Minter}},
		{Account: stakingtypes.BondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
		{Account: stakingtypes.NotBondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
		{Account: govtypes.ModuleName, Permissions: []string{authtypes.Burner}},
		{Account: ibctransfertypes.ModuleName, Permissions: []string{authtypes.Minter, authtypes.Burner}},
		{Account: wasmtypes.ModuleName, Permissions: []string{authtypes.Burner}},
		{Account: vestingstypes.ModuleName},
		{Account: icatypes.ModuleName},
		{Account: interchainqueriestypes.ModuleName},
		{Account: feetypes.ModuleName},
	}

	// TODO
	// blocked account addresses
	blockAccAddrs = []string{
		// authtypes.FeeCollectorName,
		// distrtypes.ModuleName,
		// minttypes.ModuleName,
		// stakingtypes.BondedPoolName,
		// stakingtypes.NotBondedPoolName,
		// nft.ModuleName,
		// We allow the following module accounts to receive funds:
		// govtypes.ModuleName
	}

	// application configuration (used by depinject)
	AppConfig = depinject.Configs(appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: runtime.ModuleName,
				Config: appconfig.WrapAny(&runtimev1alpha1.Module{
					AppName: "nolusd",
					// NOTE: upgrade module is required to be prioritized
					PreBlockers: []string{
						upgradetypes.ModuleName,
					},
					// During begin block slashing happens after distr.BeginBlocker so that
					// there is nothing left over in the validator fee pool, so as to keep the
					// CanWithdrawInvariant invariant.
					// NOTE: staking module is required if HistoricalEntries param > 0
					BeginBlockers: []string{
						capabilitytypes.ModuleName,
						minttypes.ModuleName,
						distrtypes.ModuleName,
						slashingtypes.ModuleName,
						evidencetypes.ModuleName,
						stakingtypes.ModuleName,
						ibcexported.ModuleName,
						genutiltypes.ModuleName,
						banktypes.ModuleName,
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
						interchainqueriestypes.ModuleName,
						contractmanagermoduletypes.ModuleName,
						wasmtypes.ModuleName,
						feetypes.ModuleName,
					},
					EndBlockers: []string{
						crisistypes.ModuleName,
						govtypes.ModuleName,
						stakingtypes.ModuleName,
						ibcexported.ModuleName,
						paramstypes.ModuleName,
						slashingtypes.ModuleName,
						upgradetypes.ModuleName,
						authtypes.ModuleName,
						capabilitytypes.ModuleName,
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
						interchainqueriestypes.ModuleName,
						contractmanagermoduletypes.ModuleName,
						wasmtypes.ModuleName,
						feetypes.ModuleName,
					},
					OverrideStoreKeys: []*runtimev1alpha1.StoreKeyConfig{},

					// NOTE: The genutils module must occur after staking so that pools are
					// properly initialized with tokens from genesis accounts.
					// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
					// NOTE: Capability module must occur first so that it can initialize any capabilities
					// so that other modules that want to create or claim capabilities afterwards in InitChain
					InitGenesis: []string{
						capabilitytypes.ModuleName,
						authtypes.ModuleName,
						banktypes.ModuleName,
						distrtypes.ModuleName,
						stakingtypes.ModuleName,
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
						interchainqueriestypes.ModuleName,
						interchaintxstypes.ModuleName,
						contractmanagermoduletypes.ModuleName,
						// wasm after ibc transfer
						wasmtypes.ModuleName,
						feetypes.ModuleName,
						consensusparamtypes.ModuleName,
					},
					// When ExportGenesis is not specified, the export genesis module order
					// is equal to the init genesis order
					// ExportGenesis: []string{},
					// Uncomment if you want to set a custom migration order here.
					// OrderMigrations: []string{},
				}),
			},
			// TODO: capability module config : ? -> generate .pulsar. files for the modules below
			// {
			// 	Name:   capabilitytypes.ModuleName,
			// 	Config: appconfig.WrapAny(&capabilitymodulev1.Module{}),
			// },
			{
				Name: authtypes.ModuleName,
				Config: appconfig.WrapAny(&authmodulev1.Module{
					Bech32Prefix:             appparams.Bech32PrefixAccAddr,
					ModuleAccountPermissions: moduleAccPerms,
					// By default modules authority is the governance module. This is configurable with the following:
					// Authority: "group", // A custom module authority can be set using a module name
					// Authority: "cosmos1cwwv22j5ca08ggdv9c2uky355k908694z577tv", // or a specific address
				}),
			},
			{
				Name: banktypes.ModuleName,
				Config: appconfig.WrapAny(&bankmodulev1.Module{
					BlockedModuleAccountsOverride: blockAccAddrs,
				}),
			},
			{
				Name:   distrtypes.ModuleName,
				Config: appconfig.WrapAny(&distrmodulev1.Module{}),
			},
			{
				Name: stakingtypes.ModuleName,
				Config: appconfig.WrapAny(&stakingmodulev1.Module{
					// NOTE: specifying a prefix is only necessary when using bech32 addresses
					// If not specfied, the auth Bech32Prefix appended with "valoper" and "valcons" is used by default
					Bech32PrefixValidator: appparams.Bech32PrefixValAddr,
					Bech32PrefixConsensus: appparams.Bech32PrefixConsAddr,
				}),
			},
			{
				Name:   slashingtypes.ModuleName,
				Config: appconfig.WrapAny(&slashingmodulev1.Module{}),
			},
			{
				Name:   govtypes.ModuleName,
				Config: appconfig.WrapAny(&govmodulev1.Module{}),
			},
			{
				Name:   minttypes.ModuleName,
				Config: appconfig.WrapAny(&mintmodulev1.Module{}),
			},
			{
				Name:   crisistypes.ModuleName,
				Config: appconfig.WrapAny(&crisismodulev1.Module{}),
			},
			// TODO: generate .pulsar. files for the modules below
			// {
			// 	Name:   taxmoduletypes.ModuleName,
			// 	Config: appconfig.WrapAny(&taxmoduletypes.Module{}),
			// },
			// {
			// 	Name:   vestingstypes.ModuleName,
			// 	Config: appconfig.WrapAny(&vestingstypes.Module{}),
			// },
			// {
			// 	Name:   ibcexported.ModuleName,
			// 	Config: appconfig.WrapAny(&ibcexported.Module{}),
			// },
			{
				Name:   genutiltypes.ModuleName,
				Config: appconfig.WrapAny(&genutilmodulev1.Module{}),
			},
			{
				Name:   evidencetypes.ModuleName,
				Config: appconfig.WrapAny(&evidencemodulev1.Module{}),
			},
			{
				Name:   feegrant.ModuleName,
				Config: appconfig.WrapAny(&feegrantmodulev1.Module{}),
			},
			{
				Name:   authz.ModuleName,
				Config: appconfig.WrapAny(&authzmodulev1.Module{}),
			},
			{
				Name:   paramstypes.ModuleName,
				Config: appconfig.WrapAny(&paramsmodulev1.Module{}),
			},
			{
				Name:   upgradetypes.ModuleName,
				Config: appconfig.WrapAny(&upgrademodulev1.Module{}),
			},
			// TODO: generate .pulsar. files for the modules below
			// {
			// 	Name:   ibctransfertypes.ModuleName,
			// 	Config: appconfig.WrapAny(&ibctransfertypes.Module{}),
			// },
			// {
			// 	Name:   icatypes.ModuleName,
			// 	Config: appconfig.WrapAny(&icatypes.Module{}),
			// },
			// {
			// 	Name:   interchainqueriestypes.ModuleName,
			// 	Config: appconfig.WrapAny(&interchainqueriestypes.Module{}),
			// },
			// {
			// 	Name:   interchaintxstypes.ModuleName,
			// 	Config: appconfig.WrapAny(&interchaintxstypes.Module{}),
			// },
			// {
			// 	Name:   contractmanagermoduletypes.ModuleName,
			// 	Config: appconfig.WrapAny(&contractmanagermoduletypes.Module{}),
			// },
			// {
			// 	Name:   wasmtypes.ModuleName,
			// 	Config: appconfig.WrapAny(&wasmtypes.Module{}),
			// },
			// {
			// 	Name:   feetypes.ModuleName,
			// 	Config: appconfig.WrapAny(&feetypes.Module{}),
			// },
			// {
			// 	Name:   consensusparamtypes.ModuleName,
			// 	Config: appconfig.WrapAny(&consensusparamtypes.Module{}),
			// },
		},
	}),
		depinject.Supply(
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
				govtypes.ModuleName: gov.NewAppModuleBasic(
					[]govclient.ProposalHandler{
						paramsclient.ProposalHandler,
					},
				),
			},
		))
)
