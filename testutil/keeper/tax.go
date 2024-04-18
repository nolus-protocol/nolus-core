package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	metrics "cosmossdk.io/store/metrics"
	sdktypes "cosmossdk.io/store/types"
	storetypes "cosmossdk.io/store/types"
	mock_types "github.com/Nolus-Protocol/nolus-core/testutil/mocks/tax/types"
	"github.com/Nolus-Protocol/nolus-core/x/tax/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	protov2 "google.golang.org/protobuf/proto"
)

func TaxKeeper(t testing.TB, isCheckTx bool, gasPrices sdk.DecCoins) (*keeper.Keeper, sdk.Context, *mock_types.MockWasmKeeper) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
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
	err := k.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return &k, ctx, mockWasmKeeper
}

type MockFeeTx struct {
	Msgs []sdk.Msg
	Gas  uint64
	Fee  sdk.Coins
}

// TODO:
func (m MockFeeTx) GetMsgsV2() ([]protov2.Message, error) {
	return []protov2.Message{}, nil // this is a hack for tests
}

func (m MockFeeTx) GetMsgs() []sdk.Msg {
	return m.Msgs
}

// func (m MockFeeTx) ValidateBasic() error {
// 	// Implement your basic validation logic here or return nil if not needed for the test.
// 	return nil
// }

func (m MockFeeTx) GetGas() uint64 {
	return m.Gas
}

func (m MockFeeTx) GetFee() sdk.Coins {
	return m.Fee
}

func (m MockFeeTx) FeePayer() []byte {
	return []byte{}
}

func (m MockFeeTx) FeeGranter() []byte {
	return []byte{}
}
