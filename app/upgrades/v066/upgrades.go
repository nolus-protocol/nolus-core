package v066

import (
	"context"
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	tmtypes "github.com/cometbft/cometbft/types"
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

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm) //nolint:contextcheck
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))

		// Properly register consensus params. In the process, change block max bytes params:
		defaultConsensusParams := tmtypes.DefaultConsensusParams().ToProto()
		defaultConsensusParams.Block.MaxBytes = 4000000 // previously 22020096
		defaultConsensusParams.Block.MaxGas = 100000000 // previously 100000000
		err = keepers.ConsensusParamsKeeper.ParamsStore.Set(ctx, defaultConsensusParams)
		if err != nil {
			return nil, err
		}
		return vm, nil
	}
}
