package v086

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	solanacarriertypes "github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// CreateUpgradeHandler for v0.8.6 introduces the x/solanacarrier module, whose
// store this upgrade adds. RunMigrations already runs InitGenesis for a module
// absent from the version map, which seeds the default parameters; the explicit
// SetParams that follows pins the fee payer this release ships regardless of how
// the module entered the version map, so the value is never left to inference.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm) //nolint:contextcheck
		if err != nil {
			return vm, err
		}

		if err := keepers.SolanaCarrierKeeper.SetParams(ctx, solanacarriertypes.DefaultParams()); err != nil {
			return vm, err
		}
		ctx.Logger().Info("Set default solanacarrier params")

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
