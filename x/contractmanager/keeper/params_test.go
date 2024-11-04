package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/contractmanager/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.ContractManagerKeeper(t, nil)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	actualParams, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.EqualValues(t, params, actualParams)
}
