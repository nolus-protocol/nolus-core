package keeper

import (
	"testing"

	"cosmossdk.io/log/v2"
	db2 "github.com/cosmos/cosmos-db"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/store/v2"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	keeper "github.com/Nolus-Protocol/nolus-core/x/transfer/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/transfer/types"
)

func TransferKeeper(
	t testing.TB,
	managerKeeper types.WasmKeeper,
	channelKeeper types.ChannelKeeper,
	authKeeper types.AccountKeeper,
) (*keeper.KeeperTransferWrapper, sdk.Context, *storetypes.KVStoreKey) {
	storeKey := storetypes.NewKVStoreKey(transfertypes.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	memStoreKey := storetypes.NewMemoryStoreKey("mem_" + transfertypes.StoreKey)

	db := db2.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	addrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	k := keeper.NewKeeper(
		cdc,
		addrCodec,
		storeService,
		channelKeeper,
		nil, // msgRouter
		authKeeper,
		nil, // bankKeeper
		managerKeeper,
		"authority",
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, transfertypes.DefaultParams())

	return &k, ctx, storeKey
}
