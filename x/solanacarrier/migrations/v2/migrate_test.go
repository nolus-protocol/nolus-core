package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	types "github.com/Nolus-Protocol/nolus-core/solanacarrier"
	"github.com/Nolus-Protocol/nolus-core/x/solanacarrier/keeper"
	v2 "github.com/Nolus-Protocol/nolus-core/x/solanacarrier/migrations/v2"
)

type fixture struct {
	ctx      sdk.Context
	storeKey *storetypes.KVStoreKey
	keeper   keeper.Keeper
}

func newFixture(t *testing.T) fixture {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	return fixture{
		ctx:      sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger()),
		storeKey: storeKey,
		keeper: keeper.NewKeeper(
			codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
			runtime.NewKVStoreService(storeKey),
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		),
	}
}

func v1ParamsBytes(feePayer []byte) []byte {
	return append([]byte{0x0a, byte(len(feePayer))}, feePayer...) // field 1, fee_payer, alone
}

func TestMigrateAddsTheAllowlistAndKeepsTheFeePayer(t *testing.T) {
	// Off curve and not the default fee payer, so keeping it is told apart from a reset.
	feePayer := []byte{
		0x4a, 0x7c, 0x9e, 0x0b, 0x4d, 0xdd, 0x06, 0x9c,
		0x64, 0x31, 0x08, 0x5d, 0x1c, 0x0b, 0x1e, 0x2f,
		0x2b, 0x5b, 0x3d, 0x9a, 0x84, 0x1f, 0x5e, 0x11,
		0xa2, 0xc7, 0x6d, 0x38, 0x90, 0x44, 0x1b, 0x00,
	}
	f := newFixture(t)
	f.ctx.KVStore(f.storeKey).Set(types.ParamsKey, v1ParamsBytes(feePayer))

	require.NoError(t, v2.Migrate(f.ctx, f.keeper))

	got, err := f.keeper.GetParams(f.ctx)
	require.NoError(t, err)
	require.Equal(t, types.Params{FeePayer: feePayer, AllowedPrograms: v2.AllowedPrograms()}, got)
}

func TestMigrateFailsWithoutStoredParams(t *testing.T) {
	f := newFixture(t)

	err := v2.Migrate(f.ctx, f.keeper)
	require.ErrorIs(t, err, types.ErrParamsNotFound)
	require.Nil(t, f.ctx.KVStore(f.storeKey).Get(types.ParamsKey), "a failed migration must write nothing")
}

func TestMigrateRefusesInvalidStoredFeePayer(t *testing.T) {
	f := newFixture(t)
	stored := v1ParamsBytes([]byte{0x01, 0x02})
	f.ctx.KVStore(f.storeKey).Set(types.ParamsKey, stored)

	err := v2.Migrate(f.ctx, f.keeper)
	require.ErrorIs(t, err, types.ErrInvalidFeePayer)
	require.Equal(t, stored, f.ctx.KVStore(f.storeKey).Get(types.ParamsKey), "a failed migration must write nothing")
}

func TestMigratedAllowlistEqualsTheDefaultParams(t *testing.T) {
	require.Equal(t, types.DefaultParams().AllowedPrograms, v2.AllowedPrograms())
}
