package interchaintxs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

// InitGenesis initializes the interchaintxs module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the interchaintxs module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var err error
	genesis := types.DefaultGenesis()
	genesis.Params, err = k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return genesis
}
