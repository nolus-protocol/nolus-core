package solanacarrier

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/Nolus-Protocol/nolus-core/solanacarrier"
	"github.com/Nolus-Protocol/nolus-core/x/solanacarrier/keeper"
)

// InitGenesis initializes the solanacarrier module's state from a provided
// genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the solanacarrier module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{Params: params}
}
