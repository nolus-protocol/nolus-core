package contractmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the failure
	for _, elem := range genState.FailuresList {
		k.AddContractFailure(ctx, elem.Address, elem.SudoPayload, elem.Error)
	}
	// this line is used by starport scaffolding # genesis/module/init
	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var err error
	genesis := types.DefaultGenesis()
	genesis.Params, err = k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	genesis.FailuresList = k.GetAllFailures(ctx)

	return genesis
}
