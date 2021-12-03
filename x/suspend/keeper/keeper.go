package keeper

import (
	"fmt"
	types2 "gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey sdk.StoreKey
		memKey   sdk.StoreKey
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,

) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		memKey:   memKey,
	}
}

// GetParams todo split to multiple methods
func (k Keeper) GetParams(ctx sdk.Context) (state types2.GenesisState) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types2.GenesisStateKey)
	if b == nil {
		panic("suspend stored state must not have been nil")
	}

	k.cdc.MustUnmarshal(b, &state)
	return
}

func (k Keeper) AddProceeds(ctx sdk.Context, delta bool) {
	genState := k.GetParams(ctx)
	genState.Suspend = delta
	k.Logger(ctx).Info(fmt.Sprintf("New Suspend proceeds state: %s", genState.Suspend))
	k.SetParams(ctx, genState)
}

// SetParams stores the genesis state. Needs a refactor to store parameters as separate values
func (k Keeper) SetParams(ctx sdk.Context, genState types2.GenesisState) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&genState)
	store.Set(types2.GenesisStateKey, b)
}

func (k Keeper) SetNodeSuspend(ctx sdk.Context, msgChange *types2.MsgChangeSuspend) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(msgChange)
	store.Set(types2.SuspendStateKey, b)
}

func (k Keeper) IsNodeSuspend(ctx sdk.Context) (suspend types2.MsgChangeSuspend) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types2.SuspendStateKey)
	k.cdc.MustUnmarshal(b, &suspend)
	return suspend
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types2.ModuleName))
}
