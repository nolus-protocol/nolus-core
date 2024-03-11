package v03

import (
	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/app/params"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	icqtypes "github.com/neutron-org/neutron/v2/x/interchainqueries/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
	codec codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Upgrade handler execution...")
		ctx.Logger().Info("Running migrations")
		interchainQueriesParams := icqtypes.Params{
			QuerySubmitTimeout:  uint64(1036800),
			QueryDeposit:        sdk.NewCoins(sdk.NewCoin(params.BaseCoinUnit, sdkmath.NewInt(1000000))),
			TxQueryRemovalLimit: uint64(10000),
		}
		err := keepers.InterchainQueriesKeeper.SetParams(ctx, interchainQueriesParams)
		if err != nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
