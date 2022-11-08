package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "gitlab-nomo.credissimo.net/nomo/nolus-core/testutil/keeper"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.TaxKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
	require.EqualValues(t, params.FeeRate, k.FeeRate(ctx))
	require.EqualValues(t, params.FeeCaps, k.FeeCaps(ctx))
	require.EqualValues(t, params.ContractAddress, k.ContractAddress(ctx))
	require.EqualValues(t, params.FeeDenoms, k.FeeDenoms(ctx))
}
