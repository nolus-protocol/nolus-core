package app

import (
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

	app.registerUpgradeV1_43(upgradeInfo)
}

// performs upgrade from v0.1.39 -> v0.1.43
func (app *App) registerUpgradeV1_43(_ storetypes.UpgradeInfo) {
	const UpgradeV1_43Plan = "v0.1.43"
	app.UpgradeKeeper.SetUpgradeHandler(UpgradeV1_43Plan, func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", UpgradeV1_43Plan)
		return fromVM, nil
	})
}
