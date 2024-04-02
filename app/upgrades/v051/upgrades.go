package v051

import (
	"context"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
	codec codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		// create ICS27 Host submodule params
		hostParams := icahosttypes.Params{
			HostEnabled:   true,
			AllowMessages: []string{},
		}

		// initialize ICS27 module
		icamodule, correctTypecast := mm.Modules[icatypes.ModuleName].(ica.AppModule)
		if !correctTypecast {
			panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
		}

		// Default icacontroller parameters keep our icacontroller enabled
		// The reason to call InitModule here is because we want to enable host functionality on nolus, which was disabled before
		// InitModule sets the params for host, binds the hostKeeper to port `icahost` and claims port capability
		icamodule.InitModule(ctx, icacontrollertypes.DefaultParams(), hostParams)

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
