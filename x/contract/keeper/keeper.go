package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/Nolus-Protocol/nolus-core/x/contract/types"
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
) *Keeper {
	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
	}
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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
func (k Keeper) IBCSendPacketCallback(
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
func (k Keeper) IBCOnAcknowledgementPacketCallback(
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
func (k Keeper) IBCOnTimeoutPacketCallback(
	cachedCtx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	contractAddress,
	packetSenderAddress string,
) error {
	return nil
	// TOOD
}

// IBCReceivePacketCallback is called in the destination chain when a packet acknowledgement is written.
// The contract is expected to handle the callback within the user defined gas limit, and handle any errors,
// out of gas, or panics gracefully.
// This entry point is called with a cached context. If an error is returned, then the changes in
// this context will not be persisted, but the packet lifecycle will not be blocked.
func (k Keeper) IBCReceivePacketCallback(
	cachedCtx sdk.Context,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
	contractAddress string,
) error {
	return nil
	// TODO
}
