package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	sdktypes "cosmossdk.io/store/types"
	storetypes "cosmossdk.io/store/types"
	mock_types "github.com/Nolus-Protocol/nolus-core/testutil/mocks/tax/types"
	"github.com/Nolus-Protocol/nolus-core/x/tax/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	tmdb "github.com/cometbft/cometbft-db"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	_ = k.SetParams(ctx, types.DefaultParams())

	return &k, ctx, mockWasmKeeper
}

type MockFeeTx struct {
	Msgs []sdk.Msg
	Gas  uint64
	Fee  sdk.Coins
}

func (m MockFeeTx) GetMsgs() []sdk.Msg {
	return m.Msgs
}

func (m MockFeeTx) ValidateBasic() error {
	// Implement your basic validation logic here or return nil if not needed for the test.
	return nil
}

func (m MockFeeTx) GetGas() uint64 {
	return m.Gas
}

func (m MockFeeTx) GetFee() sdk.Coins {
	return m.Fee
}

func (m MockFeeTx) FeePayer() sdk.AccAddress {
	return sdk.AccAddress{}
}

func (m MockFeeTx) FeeGranter() sdk.AccAddress {
	return sdk.AccAddress{}
}
