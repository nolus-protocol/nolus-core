package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	sdktypes "cosmossdk.io/store/types"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey sdktypes.StoreKey
	memKey   sdktypes.StoreKey

	wasmKeeper types.WasmKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdktypes.StoreKey,
	wasmKeeper types.WasmKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		wasmKeeper: wasmKeeper,
		authority:  authority,
	}
}

// GetAuthority returns the x/tax module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
