package transfer

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"

	wrappedtypes "github.com/Nolus-Protocol/nolus-core/x/transfer/types"
)

// KeeperTransferWrapper is a wrapper for original ibc keeper to override response for "Transfer" method.
type KeeperTransferWrapper struct {
	keeper.Keeper
	channelKeeper wrappedtypes.ChannelKeeper
	SudoKeeper    wrappedtypes.WasmKeeper
}

func (k KeeperTransferWrapper) Transfer(goCtx context.Context, msg *wrappedtypes.MsgTransfer) (*wrappedtypes.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		k.Logger(ctx).Debug("Transfer: failed to parse sender address", "sender", msg.Sender)
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Sender)
	}

	isContract := k.SudoKeeper.HasContractInfo(ctx, senderAddr)

	if err := msg.Validate(isContract); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgTransfer")
	}

	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, msg.SourcePort, msg.SourceChannel)
	if !found {
		return nil, errors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", msg.SourcePort, msg.SourceChannel,
		)
	}

	transferMsg := types.NewMsgTransfer(msg.SourcePort, msg.SourceChannel, msg.Token, msg.Sender, msg.Receiver, msg.TimeoutHeight, msg.TimeoutTimestamp, msg.Memo)
	if _, err := k.Keeper.Transfer(goCtx, transferMsg); err != nil {
		return nil, err
	}

	return &wrappedtypes.MsgTransferResponse{
		SequenceId: sequence,
		Channel:    msg.SourceChannel,
	}, nil
}

func (k KeeperTransferWrapper) UpdateParams(goCtx context.Context, msg *wrappedtypes.MsgUpdateParams) (*wrappedtypes.MsgUpdateParamsResponse, error) {
	newMsg := &types.MsgUpdateParams{
		Signer: msg.Signer,
		Params: msg.Params,
	}
	if _, err := k.Keeper.UpdateParams(goCtx, newMsg); err != nil {
		return nil, err
	}

	return &wrappedtypes.MsgUpdateParamsResponse{}, nil
}

// NewKeeper creates a new IBC transfer Keeper(KeeperTransferWrapper) instance.
func NewKeeper(
	cdc codec.BinaryCodec, key store.KVStoreService, paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper wrappedtypes.ChannelKeeper, msgServiceRouter *baseapp.MsgServiceRouter,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, sudoKeeper wrappedtypes.WasmKeeper, authority string,
) KeeperTransferWrapper {
	return KeeperTransferWrapper{
		channelKeeper: channelKeeper,
		Keeper: keeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, msgServiceRouter,
			authKeeper, bankKeeper, authority),
		SudoKeeper: sudoKeeper,
	}
}
