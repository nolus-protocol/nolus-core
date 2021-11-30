package suspend

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/keeper"
	types2 "gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types2.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types2.GenesisState {
	genState := k.GetParams(ctx)
	genesis := types2.NewGenesis(genState.FeeRate, genState.FeeCaps, genState.FeeProceeds)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
