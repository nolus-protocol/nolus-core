package v087

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"
)

// CreateUpgradeHandler for v0.8.7 adds no stores. It runs the module
// migrations, which move x/solanacarrier to consensus version 2: its stored
// params keep their fee payer and gain the program allowlist (ComputeBudget and
// Lighthouse, at most four instructions each).
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ *keepers.AppKeepers,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm) //nolint:contextcheck
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
