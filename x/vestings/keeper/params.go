package keeper

import (
	"github.com/Nolus-Protocol/nolus-core/x/vestings/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of vestings module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	// Params are always empty
	return types.NewParams()
}

// SetParams sets the total set of vestings module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
