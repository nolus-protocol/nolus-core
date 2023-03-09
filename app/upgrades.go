package app

import (
	"encoding/json"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	contractmanagermoduletypes "github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/neutron-org/neutron/x/feerefunder"
	feeRefunderTypes "github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/neutron-org/neutron/x/interchainqueries"
	interchainqueriestypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	"github.com/neutron-org/neutron/x/interchaintxs"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

func (app *App) RegisterUpgradeHandlers() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	app.registerUpgradeV1_43(upgradeInfo)
	app.registerUpgradeV1_44(upgradeInfo)
	app.registerUpgradeV2_0(upgradeInfo)
	app.registerUpgradeV2_1_testnet(upgradeInfo)
}

// performs upgrade from v0.1.39 -> v0.1.43.
func (app *App) registerUpgradeV1_43(_ storetypes.UpgradeInfo) {
	const UpgradeV1_43Plan = "v0.1.43"
	app.UpgradeKeeper.SetUpgradeHandler(UpgradeV1_43Plan, func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", UpgradeV1_43Plan)
		return fromVM, nil
	})
}

// performs upgrade from v0.1.43 -> v0.1.44.
func (app *App) registerUpgradeV1_44(_ storetypes.UpgradeInfo) {
	const UpgradeV1_44Plan = "v0.1.44"
	app.UpgradeKeeper.SetUpgradeHandler(UpgradeV1_44Plan, func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", UpgradeV1_44Plan)
		return fromVM, nil
	})
}

// performs upgrade from v0.1.43 -> v0.2.0.
func (app *App) registerUpgradeV2_0(_ storetypes.UpgradeInfo) {
	const UpgradeV2_0Plan = "v0.2.0"
	app.UpgradeKeeper.SetUpgradeHandler(UpgradeV2_0Plan, func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", UpgradeV2_0Plan)
		return fromVM, nil
	})
}

func (app *App) createUpgradeHandlerTestnet(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", "v0.2.1")
		appCodec := app.appCodec
		// Register the consensus version in the version map
		// to avoid the SDK from triggering the default
		// InitGenesis function.
		fromVM["interchainqueries"] = interchainqueries.AppModule{}.ConsensusVersion()

		// Make custom genesis state and run InitGenesis for interchainqueries
		interchainQueriesCustomGenesis := interchainqueriestypes.GenesisState{
			Params: interchainqueriestypes.Params{
				QuerySubmitTimeout: 1036800,
				QueryDeposit:       sdk.NewCoins(sdk.NewCoin("unls", sdk.NewInt(1000000))),
			},
			RegisteredQueries: []*interchainqueriestypes.RegisteredQuery{},
		}
		interchainQueriesCustomGenesisJSON, err := json.Marshal(interchainQueriesCustomGenesis)
		if err != nil {
			return nil, err
		}
		app.mm.Modules["interchainqueries"].InitGenesis(ctx, appCodec, interchainQueriesCustomGenesisJSON)

		// Register the consensus version in the version map
		// to avoid the SDK from triggering the default
		// InitGenesis function.
		fromVM["interchaintxs"] = interchaintxs.AppModule{}.ConsensusVersion()

		// Make custom genesis state and run InitGenesis for interchaintxs
		interchainTxsCustomGenesis := interchaintxstypes.GenesisState{
			Params: interchaintxstypes.Params{
				MsgSubmitTxMaxMessages: 16,
			},
		}
		interchainTxsCustomGenesisJSON, err := json.Marshal(interchainTxsCustomGenesis)
		if err != nil {
			return nil, err
		}
		app.mm.Modules["interchaintxs"].InitGenesis(ctx, appCodec, interchainTxsCustomGenesisJSON)

		// Register the consensus version in the version map
		// to avoid the SDK from triggering the default
		// InitGenesis function.
		fromVM[feeRefunderTypes.ModuleName] = feerefunder.AppModule{}.ConsensusVersion()

		// Make custom genesis state and run InitGenesis for interchaintxs
		feeRefunderCustomGenesis := feeRefunderTypes.GenesisState{
			Params: feeRefunderTypes.Params{
				MinFee: feeRefunderTypes.Fee{
					AckFee: sdk.Coins{
						sdk.NewCoin("unls", sdk.NewInt(1)),
					},
					TimeoutFee: sdk.Coins{
						sdk.NewCoin("unls", sdk.NewInt(1)),
					},
				},
			},
		}
		feeRefunderCustomGenesisJSON, err := json.Marshal(feeRefunderCustomGenesis)
		if err != nil {
			return nil, err
		}
		app.mm.Modules[feeRefunderTypes.ModuleName].InitGenesis(ctx, appCodec, feeRefunderCustomGenesisJSON)

		ctx.Logger().Info("Running migrations")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	}
}

func (app *App) registerUpgradeV2_1_testnet(_ storetypes.UpgradeInfo) {
	testnetUpgrade := upgrades.Upgrade{
		UpgradeName:          "v0.2.1",
		CreateUpgradeHandler: app.createUpgradeHandlerTestnet,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{
				contractmanagermoduletypes.ModuleName,
				feeRefunderTypes.ModuleName,
			},
		},
	}
	app.UpgradeKeeper.SetUpgradeHandler(
		testnetUpgrade.UpgradeName,
		testnetUpgrade.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			&app.AppKeepers,
		),
	)
}
