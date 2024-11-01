package mint

import (
	"github.com/Nolus-Protocol/nolus-core/x/mint/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, data *types.GenesisState) {
	err := keeper.SetMinter(ctx, data.Minter)
	if err != nil {
		ctx.Logger().Error("error setting minter", "error", err)
	}

	err = keeper.SetParams(ctx, data.Params)
	if err != nil {
		ctx.Logger().Error("error setting mint params", "error", err)
	}
	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	minter := keeper.GetMinter(ctx)
	params, err := keeper.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(minter, params)
}
