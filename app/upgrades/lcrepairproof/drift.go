package lcrepairproof

import (
	"context"
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// CreateDriftUpgradeHandler rewrites MaxClockDrift on the target client. MaxClockDrift is
// fixed at client creation and no IBC message updates it, so an upgrade handler is the only
// way to correct one that was set wrong.
func CreateDriftUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	k *keepers.AppKeepers,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		clientID, cs, err := loadTargetClient(ctx, k, DriftUpgradeName)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf(
			"%s: client %s MaxClockDrift %s -> %s",
			DriftUpgradeName, clientID, cs.MaxClockDrift, NewMaxClockDrift,
		))

		cs.MaxClockDrift = NewMaxClockDrift
		k.IBCKeeper.ClientKeeper.SetClientState(ctx, clientID, cs)

		vm, err = mm.RunMigrations(ctx, configurator, vm) //nolint:contextcheck
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", DriftUpgradeName))
		return vm, nil
	}
}
