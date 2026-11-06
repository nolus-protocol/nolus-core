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
var lighthouseProgram = []byte{
	0x04, 0xdf, 0xad, 0x79, 0x62, 0xff, 0xb1, 0xdd,
	0x92, 0x5d, 0x0a, 0x9f, 0xb5, 0xe6, 0xd0, 0x0c,
	0xe6, 0x19, 0x5b, 0xa8, 0xbb, 0x3a, 0x91, 0xfd,
	0x07, 0xef, 0x98, 0x60, 0xc5, 0xe9, 0x7b, 0xb8,
}

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

func TestGetParamsReturnsStoredValue(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)

	got, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), got)

	updated := types.Params{
		FeePayer:        altFeePayer,
		AllowedPrograms: []types.AllowedProgram{{ProgramId: lighthouseProgram, MaxInstructions: 1}},
	}
	require.NoError(t, k.SetParams(ctx, updated))

	got, err = k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, updated, got)
}

func TestGetParamsErrorsWhenParamsUnset(t *testing.T) {
	k, ctx := testkeeper.SolanaCarrierKeeper(t, nil)

	_, err := k.GetParams(ctx)
	require.ErrorIs(t, err, types.ErrParamsNotFound)
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

// GetParams is reached from the sign-mode handler, whose context need not carry
// an sdk.Context; the store service would panic on one, so the keeper guards first.
func TestGetParamsErrorsOnNonSDKContext(t *testing.T) {
	params := types.DefaultParams()
	k, _ := testkeeper.SolanaCarrierKeeper(t, &params)

	require.NotPanics(t, func() {
		_, err := k.GetParams(context.Background())
		require.ErrorIs(t, err, types.ErrNoSDKContext)
	})
}

func TestGetParamsReadsThroughAWrappedSDKContext(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)

	got, err := k.GetParams(context.WithValue(context.Background(), sdk.SdkContextKey, ctx))
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), got)
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

func TestUpdateParamsRejectsInvalidAllowlistAndKeepsStoredParams(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)
	msgServer := keeper.NewMsgServerImpl(*k)

	invalid := types.DefaultParams()
	invalid.AllowedPrograms = append(invalid.AllowedPrograms, types.AllowedProgram{
		ProgramId:       invalid.AllowedPrograms[0].ProgramId,
		MaxInstructions: 1,
	})

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    invalid,
	})
	require.ErrorIs(t, err, types.ErrInvalidAllowedPrograms)

	stored, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), stored)
}

func TestUpdateParamsReplacesTheWholeAllowlist(t *testing.T) {
	params := types.DefaultParams()
	k, ctx := testkeeper.SolanaCarrierKeeper(t, &params)
	msgServer := keeper.NewMsgServerImpl(*k)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    types.Params{FeePayer: types.DefaultFeePayer},
	})
	require.NoError(t, err)

	stored, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Empty(t, stored.AllowedPrograms, "params omitting allowed_programs leave the Memo instruction alone allowed")
}
