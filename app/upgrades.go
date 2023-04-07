package app

import (
	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	"github.com/cosmos/cosmos-sdk/store/iavl"
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

func (app *App) registerUpgrade(_ storetypes.UpgradeInfo) {
	testnetUpgrade := upgrades.Upgrade{
		UpgradeName:          "v0.2.2-store-fix",
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

func (app *App) createUpgradeHandlerTestnet(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution", "name", "v0.2.2-store-fix")
		// Get the Commit multistore
		// cms := app.BaseApp.CommitMultiStore()

		// Get the underlying iavl Stores for the contractmanager and feerefunder modules
		// contractManagerStore := cms.GetCommitKVStore(app.GetKey(contractmanagertypes.StoreKey)).(*iavl.Store)
		// feeRefunderStore := cms.GetCommitKVStore(app.GetKey(feerefundertypes.StoreKey)).(*iavl.Store)

		// We found this issue thanks to a code change introduced in the cosmos-sdk v0.45.12
		// https://github.com/cosmos/gaia/issues/2313

		// Export(at latest commit height) and import the store at the latest block height
		// We do this because we didn't use a custom store loader
		// on the upgrade where the two modules(contractmanager && feerefunder) were introduced and
		// their store versions began from height 0/1 but they should have started at the height of the upgrade
		// so right now we have a gap, other modules' stores initialized at genesis are at height X, while those two modules are behind at height X-softwareUpgradeHeight
		// err := exportAndImportStoreAtLatestHeight(ctx, contractManagerStore)
		// if err != nil {
		// 	return nil, err
		// }
		// err = exportAndImportStoreAtLatestHeight(ctx, feeRefunderStore)
		// if err != nil {
		// 	return nil, err
		// }
		return app.mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// the base purpose of this export-import is to update the store's version to the latest height.
func exportAndImportStoreAtLatestHeight(ctx sdk.Context, store *iavl.Store) error {
	exporter, err := store.Export(store.LastCommitID().Version)
	if err != nil {
		return err
	}
	defer exporter.Close()

	// If there is already version for the latest or latest-1 blocks height, then we don't do anything
	if store.VersionExists(ctx.BlockHeight()) || store.VersionExists(ctx.BlockHeight()-1) {
		ctx.Logger().Info("Version is already stored. ", "v0.2.2-store-fix")
		return nil
	}

	importer, err := store.Import(ctx.BlockHeight())
	if err != nil {
		return err
	}

	// exporter.Next() can return nil or ExportDone as the second value
	// In our case, exportDone will be nil because we know that we have at least 1 node to export, and we won't need more
	exportNode, exportDone := exporter.Next()
	if exportDone != nil {
		ctx.Logger().Debug("ExportDone is called when there is no data to export. ", " v0.2.2-store-fix")
		return nil
	}

	ctx.Logger().Info("Importing store at height/version ", ctx.BlockHeight(), " v0.2.2-store-fix")
	err = importer.Add(exportNode)
	if err != nil {
		return err
	}
	// No need to call importer.Close() as it is called internally inside the Commit() method
	err = importer.Commit()
	if err != nil {
		return err
	}

	return nil
}
