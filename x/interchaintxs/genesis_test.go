package interchaintxs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/testutil/common/nullify"
	keepertest "github.com/Nolus-Protocol/nolus-core/testutil/interchaintxs/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs"
	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.InterchainTxsKeeper(t, nil, nil, nil, nil, nil, nil, nil)
	interchaintxs.InitGenesis(ctx, *k, genesisState)
	got := interchaintxs.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
