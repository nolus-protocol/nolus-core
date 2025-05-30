package tax_test

import (
	"testing"

	keepertest "github.com/Nolus-Protocol/nolus-core/testutil/keeper"
	"github.com/Nolus-Protocol/nolus-core/testutil/nullify"
	"github.com/Nolus-Protocol/nolus-core/x/tax"
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.TaxKeeper(t, false, sdk.DecCoins{}, types.DefaultParams())
	tax.InitGenesis(ctx, *k, genesisState)
	got := tax.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
