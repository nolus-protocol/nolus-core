package keeper_test

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	types "github.com/Nolus-Protocol/nolus-core/solanacarrier"
	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/solanacarrier/keeper"
)

// altFeePayer is a second off-curve account key, used to prove a param update
// actually replaces the stored value.
var altFeePayer = []byte{
	0x4a, 0x7c, 0x9e, 0x0b, 0x4d, 0xdd, 0x06, 0x9c,
	0x64, 0x31, 0x08, 0x5d, 0x1c, 0x0b, 0x1e, 0x2f,
	0x2b, 0x5b, 0x3d, 0x9a, 0x84, 0x1f, 0x5e, 0x11,
	0xa2, 0xc7, 0x6d, 0x38, 0x90, 0x44, 0x1b, 0x00,
}

func TestSetGetParamsRoundTrip(t *testing.T) {
	k, ctx := testkeeper.SolanaCarrierKeeper(t, nil)

	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	got, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), got)
}

func TestSetParamsRejectsInvalidParams(t *testing.T) {
	k, ctx := testkeeper.SolanaCarrierKeeper(t, nil)

	err := k.SetParams(ctx, types.Params{FeePayer: []byte{0x01, 0x02}})
	require.ErrorIs(t, err, types.ErrInvalidFeePayer)
}

func TestFeePayerReturnsStoredValue(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)

	feePayer, err := k.FeePayer(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultFeePayer, feePayer)

	require.NoError(t, k.SetParams(ctx, types.Params{FeePayer: altFeePayer}))

	feePayer, err = k.FeePayer(ctx)
	require.NoError(t, err)
	require.Equal(t, altFeePayer, feePayer)
}

func TestFeePayerErrorsWhenParamsUnset(t *testing.T) {
	k, ctx := testkeeper.SolanaCarrierKeeper(t, nil)

	feePayer, err := k.FeePayer(ctx)
	require.ErrorIs(t, err, types.ErrParamsNotFound)
	require.Nil(t, feePayer)
}

// Gov routes MsgUpdateParams straight to the msg server, so SetParams itself is
// the only thing standing between a proposal and an unusable fee payer.
func TestSetParamsValidatesAndLeavesStoredParamsUntouched(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)

	onCurveMemoProgram, err := hex.DecodeString("054a535a992921064d24e87160da387c7c35b5ddbc92bb81e41fa8404105448d")
	require.NoError(t, err)

	err = k.SetParams(ctx, types.Params{FeePayer: onCurveMemoProgram})
	require.ErrorIs(t, err, types.ErrInvalidFeePayer)
	require.Contains(t, err.Error(), "on-curve")

	stored, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultFeePayer, stored.FeePayer)
}

// FeePayer is reached from the sign-mode handler, whose context need not carry an
// sdk.Context; the store service would panic on one, so the keeper guards first.
func TestFeePayerErrorsOnNonSDKContext(t *testing.T) {
	params := types.DefaultParams()
	k, _ := testkeeper.SolanaCarrierKeeper(t, &params)

	require.NotPanics(t, func() {
		feePayer, err := k.FeePayer(context.Background())
		require.ErrorIs(t, err, types.ErrNoSDKContext)
		require.Nil(t, feePayer)
	})
}

func TestFeePayerReadsThroughAWrappedSDKContext(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)

	feePayer, err := k.FeePayer(context.WithValue(context.Background(), sdk.SdkContextKey, ctx))
	require.NoError(t, err)
	require.Equal(t, types.DefaultFeePayer, feePayer)
}

func TestUpdateParamsRejectsWrongAuthority(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)
	msgServer := keeper.NewMsgServerImpl(*k)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(types.ModuleName).String(),
		Params:    types.Params{FeePayer: altFeePayer},
	})
	require.ErrorIs(t, err, govtypes.ErrInvalidSigner)

	stored, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), stored)
}

func TestUpdateParamsPersistsWithGovAuthority(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)
	msgServer := keeper.NewMsgServerImpl(*k)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    types.Params{FeePayer: altFeePayer},
	})
	require.NoError(t, err)

	stored, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, altFeePayer, stored.FeePayer)
}

func TestUpdateParamsRejectsInvalidFeePayer(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)
	msgServer := keeper.NewMsgServerImpl(*k)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    types.Params{FeePayer: types.DefaultFeePayer[:31]},
	})
	require.ErrorIs(t, err, types.ErrInvalidFeePayer)
}

func TestQueryParams(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)

	res, err := k.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), res.Params)

	_, err = k.Params(ctx, nil)
	require.Error(t, err)
}
