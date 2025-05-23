package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

type WasmKeeper interface {
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
	Sudo(ctx context.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

type FeeRefunderKeeper interface {
	// LockFees(ctx context.Context, payer sdk.AccAddress, packetID feerefundertypes.PacketID, fee feerefundertypes.Fee) error
	// DistributeAcknowledgementFee(ctx context.Context, receiver sdk.AccAddress, packetID feerefundertypes.PacketID)
	// DistributeTimeoutFee(ctx context.Context, receiver sdk.AccAddress, packetID feerefundertypes.PacketID)
}

// ChannelKeeper defines the expected IBC channel keeper.
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	GetAllChannelsWithPortPrefix(ctx sdk.Context, portPrefix string) []channeltypes.IdentifiedChannel
	HasChannel(ctx sdk.Context, portID, channelID string) bool
}

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}
