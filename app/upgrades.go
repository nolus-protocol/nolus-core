package app

import (
	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (app *App) RegisterUpgradeHandlers() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	app.registerUpgrade(upgradeInfo)
}

func (app *App) createUpgradeHandlerTestnet(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", "v0.2.2")
		ctx.Logger().Info("Running migrations")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	}
}

// upgrade v0.2.2.
func (app *App) registerUpgrade(_ storetypes.UpgradeInfo) {
	testnetUpgrade := upgrades.Upgrade{
		UpgradeName:          "v0.2.2",
		CreateUpgradeHandler: app.createUpgradeHandlerTestnet,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{},
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
