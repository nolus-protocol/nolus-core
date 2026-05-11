package keeper

import (
	"testing"

	"cosmossdk.io/log/v2"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	db2 "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/store/v2"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/x/contractmanager"
	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
)

func NewSudoLimitWrapper(t testing.TB, cmKeeper types.ContractManagerKeeper, wasmKeeper types.WasmKeeper) (types.WasmKeeper, sdk.Context, *storetypes.KVStoreKey) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := db2.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	limitWrapper := contractmanager.NewSudoLimitWrapper(cmKeeper, wasmKeeper)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return limitWrapper, ctx, storeKey
}
