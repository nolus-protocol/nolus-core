package app

import (
	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	"github.com/cosmos/cosmos-sdk/store/iavl"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
	feerefundertypes "github.com/neutron-org/neutron/x/feerefunder/types"
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
		UpgradeName:          "v0.2.2-equalize-store-heights",
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
		ctx.Logger().Info("Upgrade handler execution v0.2.2-equalize-store-heights")
		// Get the Commit multistore
		cms := app.BaseApp.CommitMultiStore()

		// Get the underlying iavl Stores for the contractmanager and feerefunder modules
		contractManagerStore := cms.GetCommitKVStore(app.GetKey(contractmanagertypes.StoreKey)).(*iavl.Store)
		feeRefunderStore := cms.GetCommitKVStore(app.GetKey(feerefundertypes.StoreKey)).(*iavl.Store)

		// We found this issue thanks to a code change introduced in the cosmos-sdk v0.45.12
		// https://github.com/cosmos/gaia/issues/2313

		// Move store's height to latest
		// We do this because we didn't use a custom store loader
		// on the upgrade where the two modules(contractmanager && feerefunder) were introduced and
		// their store versions began from height 0/1 but they should have started at the height of the upgrade
		// so right now we have a gap, other modules' stores initialized at genesis are at height X, while those two modules are behind at height X-softwareUpgradeHeight
		err := commitStoreToLatestHeight(ctx, contractManagerStore)
		if err != nil {
			ctx.Logger().Info("Failed to fix contractManager store")
			return nil, err
		}
		err = commitStoreToLatestHeight(ctx, feeRefunderStore)
		if err != nil {
			ctx.Logger().Info("Failed to fix feeRefunder store")
			return nil, err
		}
		return app.mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// this function takes a store and commits the store state, moving it's version/height by X
// the purpose of this function is to move the height of a store(which is behind) to latest.
func commitStoreToLatestHeight(ctx sdk.Context, store *iavl.Store) error {
	// If there is already version for the latest or latest-1 blocks height, then we don't do anything
	if store.VersionExists(ctx.BlockHeight()) || store.VersionExists(ctx.BlockHeight()-1) {
		ctx.Logger().Info("Latest version is already stored, the store doesn't need fixing")
		return nil
	}

	ctx.Logger().Info("Equalizing store height...")
	for store.LastCommitID().Version < ctx.BlockHeight()-1 {
		store.Commit()
	}
	ctx.Logger().Info("Finished equalizing store height")

	return nil
}
