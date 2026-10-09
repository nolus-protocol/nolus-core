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

// CreateUpgradeHandler for v0.8.7 carries no store or parameter changes. The
// release moves the node to Cosmos SDK v0.53.8 (Nolus fork) and CometBFT
// v0.38.26, whose state-breaking fixes live in the binary; the coordinated
// upgrade height switches every validator at once, and the handler only runs
// the module migrations so the plan named "v0.8.7" resolves and is recorded.
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
