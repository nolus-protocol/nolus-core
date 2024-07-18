package keeper

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	metrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	mock_types "github.com/Nolus-Protocol/nolus-core/testutil/mocks/tax/types"
	"github.com/Nolus-Protocol/nolus-core/x/tax/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
)

func TaxKeeper(t testing.TB, isCheckTx bool, gasPrices sdk.DecCoins) (*keeper.Keeper, sdk.Context, *mock_types.MockWasmKeeper) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	ctrl := gomock.NewController(t)
	mockWasmKeeper := mock_types.NewMockWasmKeeper(ctrl)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
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

func (m MockFeeTx) GetMsgsV2() ([]protov2.Message, error) {
	return []protov2.Message{}, nil
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
