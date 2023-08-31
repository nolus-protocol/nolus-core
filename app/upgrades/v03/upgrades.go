package v03

import (
	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/app/params"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
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
			QueryDeposit:        sdk.NewCoins(sdk.NewCoin(params.BaseCoinUnit, sdk.NewInt(1000000))),
			TxQueryRemovalLimit: uint64(10000),
		}
		err := keepers.InterchainQueriesKeeper.SetParams(ctx, interchainQueriesParams)
		if err != nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
