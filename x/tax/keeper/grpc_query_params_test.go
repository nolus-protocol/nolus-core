package keeper_test

import (
	"testing"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	params.GetDefaultConfig()
	keeper, ctx, _ := testkeeper.TaxKeeper(t, false, sdk.DecCoins{})
	params := types.DefaultParams()
	_ = keeper.SetParams(ctx, params)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestParamsQueryNilRequest(t *testing.T) {
	keeper, ctx, _ := testkeeper.TaxKeeper(t, false, sdk.DecCoins{})
	params := types.DefaultParams()
	_ = keeper.SetParams(ctx, params)

	response, err := keeper.Params(ctx, nil)
	require.Error(t, err)
	require.Nil(t, response)
}
