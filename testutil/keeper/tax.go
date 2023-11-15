package keeper

import (
	"testing"

	mock_types "github.com/Nolus-Protocol/nolus-core/testutil/mocks/tax/types"
	"github.com/Nolus-Protocol/nolus-core/x/tax/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdktypes "github.com/cosmos/cosmos-sdk/store/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TaxKeeper(t testing.TB, isCheckTx bool, gasPrices sdk.DecCoins) (*keeper.Keeper, sdk.Context, *mock_types.MockWasmKeeper) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdktypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdktypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	ctrl := gomock.NewController(t)
	mockWasmKeeper := mock_types.NewMockWasmKeeper(ctrl)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		mockWasmKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, isCheckTx, log.NewNopLogger()).WithMinGasPrices(gasPrices)

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return &k, ctx, mockWasmKeeper
}
