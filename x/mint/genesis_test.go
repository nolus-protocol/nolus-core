package mint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/app/params"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/testutil/nullify"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/testutil/simapp"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/mint"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/mint/types"
)

func TestGenesis(t *testing.T) {
	params.SetAddressPrefixes()
	app, err := simapp.TestSetup()
	if err != nil {
		t.Errorf("Error while creating simapp: %v\"", err)
	}
	blockTime := time.Now()
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	ctx := app.BaseApp.NewContext(false, header).WithBlockTime(blockTime)
	minterKeeper := app.MintKeeper

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	acc := app.AccountKeeper
	mint.InitGenesis(ctx, minterKeeper, acc, &genesisState)
	got := mint.ExportGenesis(ctx, minterKeeper)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

}
