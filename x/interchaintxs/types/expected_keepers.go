package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	feerefundertypes "github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias).
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type WasmKeeper interface {
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
	Sudo(ctx context.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

type ICAControllerKeeper interface {
	GetActiveChannelID(ctx sdk.Context, connectionID, portID string) (string, bool)
	GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool)
	SetMiddlewareEnabled(ctx sdk.Context, portID, connectionID string)
}

type ICAControllerMsgServer interface {
	RegisterInterchainAccount(context.Context, *icacontrollertypes.MsgRegisterInterchainAccount) (*icacontrollertypes.MsgRegisterInterchainAccountResponse, error)
	SendTx(context.Context, *icacontrollertypes.MsgSendTx) (*icacontrollertypes.MsgSendTxResponse, error)
}

type FeeRefunderKeeper interface {
	LockFees(ctx context.Context, payer sdk.AccAddress, packetID feerefundertypes.PacketID, fee feerefundertypes.Fee) error
	DistributeAcknowledgementFee(ctx context.Context, receiver sdk.AccAddress, packetID feerefundertypes.PacketID)
	DistributeTimeoutFee(ctx context.Context, receiver sdk.AccAddress, packetID feerefundertypes.PacketID)
}

// ChannelKeeper defines the expected IBC channel keeper.
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, error)
}
