package v2

import (
	"context"

	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
)

// MigrateStore performs in-place store migrations.
// The migration rearranges removes all old failures,
// since they do not have the necessary fields packet and ack for resubmission.
func MigrateStore(ctx context.Context, storeService corestoretypes.KVStoreService) error {
	return migrateFailures(ctx, runtime.KVStoreAdapter(storeService.OpenKVStore(ctx)))
}

func migrateFailures(ctx context.Context, store storetypes.KVStore) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.Logger().Info("Migrating failures...")

	// fetch list of all old failure keys
	failureKeys := make([][]byte, 0)
	iteratorStore := prefix.NewStore(store, types.ContractFailuresKey)
	iterator := storetypes.KVStorePrefixIterator(iteratorStore, []byte{})

	for ; iterator.Valid(); iterator.Next() {
		failureKeys = append(failureKeys, iterator.Key())
	}

	err := iterator.Close()
	if err != nil {
		return err
	}

	// remove failures
	store = prefix.NewStore(store, types.ContractFailuresKey)
	for _, key := range failureKeys {
		store.Delete(key)
	}

	sdkCtx.Logger().Info("Finished migrating failures")

	return nil
}
