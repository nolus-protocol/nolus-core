package tax_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/app/params"
	keepertest "gitlab-nomo.credissimo.net/nomo/cosmzone/testutil/keeper"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/testutil/nullify"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/types"
)

func TestGenesis(t *testing.T) {
	params.SetAddressPrefixes()
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.TaxKeeper(t)
	tax.InitGenesis(ctx, *k, genesisState)
	got := tax.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
