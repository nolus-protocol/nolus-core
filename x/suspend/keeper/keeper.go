package keeper

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"

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

) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		memKey:   memKey,
	}
}

func (k Keeper) SetState(ctx sdk.Context, state types.SuspendedState) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&state)
	store.Set(types.SuspendStateKey, b)
}

func (k Keeper) ChangeSuspendedState(ctx sdk.Context, msg *types.MsgChangeSuspended) error {
	state := k.GetState(ctx)
	if len(state.AdminAddress) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "No admin address is set")
	}
	adminAcc, err := sdk.AccAddressFromBech32(state.AdminAddress)
	if err != nil {
		return err
	}
	fromAcc, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}
	if !adminAcc.Equals(fromAcc) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to change suspended state", msg.FromAddress)
	}
	state.Suspended = msg.Suspended
	state.BlockHeight = msg.BlockHeight
	k.SetState(ctx, state)
	return nil
}

func (k Keeper) GetState(ctx sdk.Context) (suspend types.SuspendedState) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SuspendStateKey)
	k.cdc.MustUnmarshal(b, &suspend)
	return suspend
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
