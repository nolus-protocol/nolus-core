package transfer

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a sudo call.
func (im IBCModule) HandleAcknowledgement(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	// TODO handle channelVersion cheking here if we want to support V2
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}
	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	senderAddress, err := sdk.AccAddressFromBech32(data.GetSender())
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to decode address from bech32: %v", err)
	}
	if !im.sudoKeeper.HasContractInfo(ctx, senderAddress) {
		return nil
	}

	msg, err := keeper.PrepareSudoCallbackMessage(packet, &ack)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal Packet/Acknowledgment: %v", err)
	}

	_, err = im.sudoKeeper.Sudo(ctx, senderAddress, msg)
	if err != nil {
		im.keeper.Logger(ctx).Debug("HandleAcknowledgement: failed to Sudo contract on packet acknowledgement", "error", err)
	}

	im.keeper.Logger(ctx).Debug("acknowledgement received", "Packet data", data, "CheckTx", ctx.IsCheckTx())

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a sudo call.
func (im IBCModule) HandleTimeout(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	// TODO handle channelVersion cheking here if we want to support V2
	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	senderAddress, err := sdk.AccAddressFromBech32(data.GetSender())
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to decode address from bech32: %v", err)
	}
	if !im.sudoKeeper.HasContractInfo(ctx, senderAddress) {
		return nil
	}

	msg, err := keeper.PrepareSudoCallbackMessage(packet, nil)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal Packet: %v", err)
	}

	_, err = im.sudoKeeper.Sudo(ctx, senderAddress, msg)
	if err != nil {
		im.keeper.Logger(ctx).Debug("HandleAcknowledgement: failed to Sudo contract on packet timeout", "error", err)
	}

	return nil
}
