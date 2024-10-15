package keeper_test

import (
	"testing"

	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/feerefunder/keeper"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.FeeKeeper(t, nil, nil)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
