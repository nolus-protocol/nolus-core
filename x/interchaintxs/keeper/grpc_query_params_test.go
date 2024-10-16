package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/interchaintxs/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.InterchainTxsKeeper(t, nil, nil, nil, nil, nil, nil, nil)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(ctx, nil)
	require.Error(t, err)
	require.Nil(t, response)

	response, err = keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}