package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		wasmKeeper   types.WasmKeeper
		authority    string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	wasmKeeper types.WasmKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		wasmKeeper:   wasmKeeper,
		authority:    authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx context.Context) log.Logger {
	c := sdk.UnwrapSDKContext(ctx)
	return c.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
