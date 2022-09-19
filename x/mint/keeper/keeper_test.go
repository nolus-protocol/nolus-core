package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/app/params"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/testutil/nullify"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/testutil/simapp"
)

func TestSetMinter(t *testing.T) {
	params.SetAddressPrefixes()
	app, err := simapp.TestSetup()
	if err != nil {
		t.Errorf("Error while creating simapp: %v\"", err)
	}
	blockTime := time.Now()
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	ctx := app.BaseApp.NewContext(false, header).WithBlockTime(blockTime)
	minterKeeper := app.MintKeeper

	got := minterKeeper.GetMinter(ctx)
	require.NotNil(t, got)

	nullify.Fill(got)
}
