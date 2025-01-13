package v069dev

import (
	"context"
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
	codec codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("[dev]Deleting proposal 281...")
		keepers.GovKeeper.DeleteProposal(ctx, 281)
		ctx.Logger().Info("[dev]Deleting proposal 282...")
		keepers.GovKeeper.DeleteProposal(ctx, 282)
		ctx.Logger().Info("[dev]Deleting proposal 283...")
		keepers.GovKeeper.DeleteProposal(ctx, 283)

		keepers.TaxKeeper.SetParams(ctx, typesv2.DefaultParams())

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm) //nolint:contextcheck
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
