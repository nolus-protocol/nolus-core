package transfer

import (
	"context"

	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	wrappedtypes "github.com/Nolus-Protocol/nolus-core/x/transfer/types"
)

// KeeperTransferWrapper is a wrapper for original ibc keeper to override response for "Transfer" method.
type KeeperTransferWrapper struct {
	keeper.Keeper
	channelKeeper wrappedtypes.ChannelKeeper
	FeeKeeper     wrappedtypes.FeeRefunderKeeper
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
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper wrappedtypes.ChannelKeeper, portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, scopedKeeper capabilitykeeper.ScopedKeeper,
	feeKeeper wrappedtypes.FeeRefunderKeeper,
	sudoKeeper wrappedtypes.WasmKeeper, authority string,
) KeeperTransferWrapper {
	return KeeperTransferWrapper{
		channelKeeper: channelKeeper,
		Keeper: keeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper,
			authKeeper, bankKeeper, scopedKeeper, authority),
		FeeKeeper:  feeKeeper,
		SudoKeeper: sudoKeeper,
	}
}

// IBCSendPacketCallback is called in the source chain when a PacketSend is executed. The
// packetSenderAddress is determined by the underlying module, and may be empty if the sender is
// unknown or undefined. The contract is expected to handle the callback within the user defined
// gas limit, and handle any errors, or panics gracefully.
// This entry point is called with a cached context. If an error is returned, then the changes in
// this context will not be persisted, and the error will be propagated to the underlying IBC
// application, resulting in a packet send failure.
//
// Implementations are provided with the packetSenderAddress and MAY choose to use this to perform
// validation on the origin of a given packet. It is recommended to perform the same validation
// on all source chain callbacks (SendPacket, AcknowledgementPacket, TimeoutPacket). This
// defensively guards against exploits due to incorrectly wired SendPacket ordering in IBC stacks.
func (k KeeperTransferWrapper) IBCSendPacketCallback(
	cachedCtx sdk.Context,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	packetData []byte,
	contractAddress,
	packetSenderAddress string,
) error {
	return nil
	// TODO
}

// IBCOnAcknowledgementPacketCallback is called in the source chain when a packet acknowledgement
// is received. The packetSenderAddress is determined by the underlying module, and may be empty if
// the sender is unknown or undefined. The contract is expected to handle the callback within the
// user defined gas limit, and handle any errors, or panics gracefully.
// This entry point is called with a cached context. If an error is returned, then the changes in
// this context will not be persisted, but the packet lifecycle will not be blocked.
//
// Implementations are provided with the packetSenderAddress and MAY choose to use this to perform
// validation on the origin of a given packet. It is recommended to perform the same validation
// on all source chain callbacks (SendPacket, AcknowledgementPacket, TimeoutPacket). This
// defensively guards against exploits due to incorrectly wired SendPacket ordering in IBC stacks.
func (k KeeperTransferWrapper) IBCOnAcknowledgementPacketCallback(
	cachedCtx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
	contractAddress,
	packetSenderAddress string,
) error {
	return nil
	// TODO
}

// IBCOnTimeoutPacketCallback is called in the source chain when a packet is not received before
// the timeout height. The packetSenderAddress is determined by the underlying module, and may be
// empty if the sender is unknown or undefined. The contract is expected to handle the callback
// within the user defined gas limit, and handle any error, out of gas, or panics gracefully.
// This entry point is called with a cached context. If an error is returned, then the changes in
// this context will not be persisted, but the packet lifecycle will not be blocked.
//
// Implementations are provided with the packetSenderAddress and MAY choose to use this to perform
// validation on the origin of a given packet. It is recommended to perform the same validation
// on all source chain callbacks (SendPacket, AcknowledgementPacket, TimeoutPacket). This
// defensively guards against exploits due to incorrectly wired SendPacket ordering in IBC stacks.
func (k KeeperTransferWrapper) IBCOnTimeoutPacketCallback(
	cachedCtx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	contractAddress,
	packetSenderAddress string,
) error {
	return nil
	// TODO
}

// IBCReceivePacketCallback is called in the destination chain when a packet acknowledgement is written.
// The contract is expected to handle the callback within the user defined gas limit, and handle any errors,
// out of gas, or panics gracefully.
// This entry point is called with a cached context. If an error is returned, then the changes in
// this context will not be persisted, but the packet lifecycle will not be blocked.
func (k KeeperTransferWrapper) IBCReceivePacketCallback(
	cachedCtx sdk.Context,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
	contractAddress string,
) error {
	return nil
	// TODO
}
