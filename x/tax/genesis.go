package tax

import (
	"github.com/Nolus-Protocol/nolus-core/x/tax/keeper"
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the tax module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		ctx.Logger().Error("failed to set tax module params", "error", err)
	}
}

// ExportGenesis returns the tax module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	genesis.Params = params

	return genesis
}
