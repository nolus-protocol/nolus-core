package v090

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"
)

// CreateUpgradeHandler for v0.9.0 runs the module migrations, then the recovery sweep configured for the
// chain, if any. It panics when the embedded lists are invalid.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak *keepers.AppKeepers,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	configs := mustParseEmbedded()
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, fmt.Errorf("%s: run module migrations: %w", UpgradeName, err)
		}

		cfg, ok := configs[ctx.ChainID()]
		if !ok {
			ctx.Logger().Info("no recovery sweep configured for this chain", "chain_id", ctx.ChainID())
		} else if err := Sweep(ctx, ak, cfg); err != nil {
			return vm, fmt.Errorf("%s: recovery sweep: %w", UpgradeName, err)
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
