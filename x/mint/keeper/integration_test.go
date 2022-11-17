package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/app/params"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/testutil/simapp"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/mint/types"
)

// returns context and an app with updated mint keeper
func TestSetAndRetrieveParamsAndMinter(t *testing.T) {
	denom := "unls"
	maxMintableNanoseconds := uint64(2000)

	params.SetAddressPrefixes()
	app, err := simapp.TestSetup()
	require.NoError(t, err)

	isCheckTx := false
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})

	app.MintKeeper.SetParams(ctx, types.NewParams(denom, maxMintableNanoseconds))
	app.MintKeeper.SetMinter(ctx, types.DefaultInitialMinter())

	require.Equal(t, denom, app.MintKeeper.GetParams(ctx).MintDenom)
	require.Equal(t, maxMintableNanoseconds, app.MintKeeper.GetParams(ctx).MaxMintableNanoseconds)
	require.Equal(t, types.DefaultInitialMinter(), app.MintKeeper.GetMinter(ctx))
}
