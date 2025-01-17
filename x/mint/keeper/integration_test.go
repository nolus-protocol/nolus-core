package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/testutil/simapp"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

// returns context and an app with updated mint keeper.
func TestSetAndRetrieveParamsAndMinter(t *testing.T) {
	denom := "unls"
	maxMintableNanoseconds := uint64(2000)

	params.GetDefaultConfig()
	app, err := simapp.TestSetup(t)
	require.NoError(t, err)

	isCheckTx := false
	ctx := app.BaseApp.NewContext(isCheckTx)

	err = app.MintKeeper.SetParams(ctx, types.NewParams(denom, sdkmath.NewUint(maxMintableNanoseconds)))
	require.NoError(t, err)

	err = app.MintKeeper.SetMinter(ctx, types.DefaultInitialMinter())
	require.NoError(t, err)

	params, err := app.MintKeeper.GetParams(ctx)
	require.NoError(t, err)

	require.Equal(t, denom, params.MintDenom)
	require.Equal(t, sdkmath.NewUint(maxMintableNanoseconds), params.MaxMintableNanoseconds)
	require.Equal(t, types.DefaultInitialMinter(), app.MintKeeper.GetMinter(ctx))
}
