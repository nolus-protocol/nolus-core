package v062

import (
	"context"
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

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

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm) //nolint:contextcheck
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(`
$$\   $$\           $$\                                       $$$$$$\      $$$$$$\       $$$$$$\  
$$$\  $$ |          $$ |                                     $$$ __$$\    $$  __$$\     $$  __$$\ 
$$$$\ $$ | $$$$$$\  $$ |$$\   $$\  $$$$$$$\       $$\    $$\ $$$$\ $$ |   $$ /  \__|    \__/  $$ |
$$ $$\$$ |$$  __$$\ $$ |$$ |  $$ |$$  _____|      \$$\  $$  |$$\$$\$$ |   $$$$$$$\       $$$$$$  |
$$ \$$$$ |$$ /  $$ |$$ |$$ |  $$ |\$$$$$$\         \$$\$$  / $$ \$$$$ |   $$  __$$\     $$  ____/ 
$$ |\$$$ |$$ |  $$ |$$ |$$ |  $$ | \____$$\         \$$$  /  $$ |\$$$ |   $$ /  $$ |    $$ |      
$$ | \$$ |\$$$$$$  |$$ |\$$$$$$  |$$$$$$$  |         \$  /   \$$$$$$  /$$\ $$$$$$  |$$\ $$$$$$$$\ 
\__|  \__| \______/ \__| \______/ \_______/           \_/     \______/ \__|\______/ \__|\________| 																									   
		
$$$$$$$$\      $$\                                                                               
$$  _____|     $$ |                                                                              
$$ |      $$$$$$$ | $$$$$$\  $$$$$$$\                                                            
$$$$$\   $$  __$$ |$$  __$$\ $$  __$$\                                                           
$$  __|  $$ /  $$ |$$$$$$$$ |$$ |  $$ |                                                          
$$ |     $$ |  $$ |$$   ____|$$ |  $$ |                                                          
$$$$$$$$\\$$$$$$$ |\$$$$$$$\ $$ |  $$ |                                                          
\________|\_______| \_______|\__|  \__|
														 
  $$\ $$\   $$\       $$$$$$$$\ $$\       
  $$ \$$ \  $$ |      $$  _____|$$ |      
$$$$$$$$$$\ $$ |      $$ |      $$ |      
\_$$  $$   |$$ |      $$$$$\    $$ |      
$$$$$$$$$$\ $$ |      $$  __|   $$ |      
\_$$  $$  _|$$ |      $$ |      $$ |      
  $$ |$$ |  $$$$$$$$\ $$ |      $$$$$$$$\ 
  \__|\__|  \________|\__|      \________|
`)

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
