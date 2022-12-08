package tax_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/app/params"
	keepertest "gitlab-nomo.credissimo.net/nomo/nolus-core/testutil/keeper"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/testutil/nullify"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"
)

func TestGenesis(t *testing.T) {
	params.SetAddressPrefixes()
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.TaxKeeper(t)
	tax.InitGenesis(ctx, *k, genesisState)
	got := tax.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
