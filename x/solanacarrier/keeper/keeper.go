package keeper

import (
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
	}
}
