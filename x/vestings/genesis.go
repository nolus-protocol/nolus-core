package alloc

import (
	"github.com/Nolus-Protocol/nolus-core/x/vestings/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/vestings/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the vestings module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the vestings module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
