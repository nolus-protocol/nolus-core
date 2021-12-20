package suspend

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/keeper"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetState(ctx, genState.State)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genState := k.GetState(ctx)
	state := types.NewSuspendedState(genState.AdminAddress, genState.Suspended, genState.BlockHeight)
	genesis := types.NewGenesisState(state)

	// this line is used by starport scaffolding # genesis/module/export
	return genesis
}
