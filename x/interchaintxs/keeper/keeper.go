package keeper

import (
	"fmt"

	"cosmossdk.io/log"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

const (
	LabelSubmitTx                  = "submit_tx"
	LabelHandleAcknowledgment      = "handle_ack"
	LabelLabelHandleChanOpenAck    = "handle_chan_open_ack"
	LabelRegisterInterchainAccount = "register_interchain_account"
	LabelHandleTimeout             = "handle_timeout"
)

type (
	Keeper struct {
		Codec                  codec.BinaryCodec
		storeKey               storetypes.StoreKey
		memKey                 storetypes.StoreKey
		channelKeeper          types.ChannelKeeper
		feeKeeper              types.FeeRefunderKeeper
		icaControllerKeeper    types.ICAControllerKeeper
		icaControllerMsgServer types.ICAControllerMsgServer
		sudoKeeper             types.WasmKeeper
		bankKeeper             types.BankKeeper
		getFeeCollectorAddr    types.GetFeeCollectorAddr
		authority              string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	channelKeeper types.ChannelKeeper,
	icaControllerKeeper types.ICAControllerKeeper,
	icaControllerMsgServer types.ICAControllerMsgServer,
	sudoKeeper types.WasmKeeper,
	feeKeeper types.FeeRefunderKeeper,
	bankKeeper types.BankKeeper,
	getFeeCollectorAddr types.GetFeeCollectorAddr,
	authority string,
) *Keeper {
	return &Keeper{
		Codec:                  cdc,
		storeKey:               storeKey,
		memKey:                 memKey,
		channelKeeper:          channelKeeper,
		icaControllerKeeper:    icaControllerKeeper,
		icaControllerMsgServer: icaControllerMsgServer,
		sudoKeeper:             sudoKeeper,
		feeKeeper:              feeKeeper,
		bankKeeper:             bankKeeper,
		getFeeCollectorAddr:    getFeeCollectorAddr,
		authority:              authority,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetAuthority() string {
	return k.authority
}
