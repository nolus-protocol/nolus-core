package keeper_test

import (
	"testing"

	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/feerefunder/keeper"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.FeeKeeper(t, nil, nil)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	if err != nil {
		panic(err)
	}

	keeperParams, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	require.EqualValues(t, params, keeperParams)
}
