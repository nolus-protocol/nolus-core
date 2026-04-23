package keeper

import (
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/keeper"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"

	contractmanagertypes "github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

var (
	ibcMeter                  = otel.Meter("github.com/Nolus-Protocol/nolus-core/x/interchaintxs")
	handleAckDuration         metric.Float64Histogram
	handleTimeoutDuration     metric.Float64Histogram
	handleChanOpenAckDuration metric.Float64Histogram
)

func init() {
	var err error
	handleAckDuration, err = ibcMeter.Float64Histogram(
		LabelHandleAcknowledgment,
		metric.WithDescription("Duration of IBC acknowledgement handling including CosmWasm sudo call"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
	handleTimeoutDuration, err = ibcMeter.Float64Histogram(
		LabelHandleTimeout,
		metric.WithDescription("Duration of IBC timeout handling including CosmWasm sudo call"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
	handleChanOpenAckDuration, err = ibcMeter.Float64Histogram(
		LabelLabelHandleChanOpenAck,
		metric.WithDescription("Duration of IBC channel open ack handling including CosmWasm sudo call"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
}

// HandleAcknowledgement passes the acknowledgement data to the appropriate contract via a sudo call.
func (k *Keeper) HandleAcknowledgement(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	// TODO - in order to support v2 check channelVersion here and decide what to do - right now we only use - ibc-gotransfertypes.V1 = "ics20-1" and ibc-goICAtypes.Version = "ics27-1"
	// So far for the ICA there is only version 1 as far as I can see. For the transfer module there is V2 but it won't be handled here
	start := time.Now()
	var contract string
	defer func() {
		handleAckDuration.Record(ctx.Context(), time.Since(start).Seconds(),
			metric.WithAttributes(attribute.String("contract", contract)),
		)
	}()
	k.Logger(ctx).Debug("Handling acknowledgement")
	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	if err != nil {
		k.Logger(ctx).Error("HandleAcknowledgement: failed to get ica owner from source port", "error", err)
		return errors.Wrap(err, "failed to get ica owner from port")
	}
	contract = icaOwner.GetContract().String()

	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		k.Logger(ctx).Error("HandleAcknowledgement: cannot unmarshal ICS-27 packet acknowledgement", "error", err)
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 packet acknowledgement: %v", err)
	}
	msg, err := keeper.PrepareSudoCallbackMessage(packet, &ack)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal Packet/Acknowledgment: %v", err)
	}

	// Actually we have only one kind of error returned from acknowledgement
	// maybe later we'll retrieve actual errors from events
	_, err = k.sudoKeeper.Sudo(ctx, icaOwner.GetContract(), msg)
	if err != nil {
		k.Logger(ctx).Debug("HandleAcknowledgement: failed to Sudo contract on packet acknowledgement", "error", err)
	}

	return nil
}

// HandleTimeout passes the timeout data to the appropriate contract via a sudo call.
// Since all ICA channels are ORDERED, a single timeout shuts down a channel.
func (k *Keeper) HandleTimeout(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	start := time.Now()
	var contract string
	defer func() {
		handleTimeoutDuration.Record(ctx.Context(), time.Since(start).Seconds(),
			metric.WithAttributes(attribute.String("contract", contract)),
		)
	}()
	k.Logger(ctx).Debug("HandleTimeout")
	icaOwner, err := types.ICAOwnerFromPort(packet.SourcePort)
	if err != nil {
		k.Logger(ctx).Error("HandleTimeout: failed to get ica owner from source port", "error", err)
		return errors.Wrap(err, "failed to get ica owner from port")
	}
	contract = icaOwner.GetContract().String()

	msg, err := keeper.PrepareSudoCallbackMessage(packet, nil)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal Packet: %v", err)
	}

	_, err = k.sudoKeeper.Sudo(ctx, icaOwner.GetContract(), msg)
	if err != nil {
		k.Logger(ctx).Debug("HandleTimeout: failed to Sudo contract on packet timeout", "error", err)
	}

	return nil
}

// HandleChanOpenAck passes the data about a successfully created channel to the appropriate contract
// (== the data about a successfully registered interchain account).
// Notice that in the case of an ICA channel - it is not yet in OPEN state here
// the last step of channel opening(confirm) happens on the host chain.
func (k *Keeper) HandleChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyChannelID,
	counterpartyVersion string,
) error {
	start := time.Now()
	var contract string
	defer func() {
		handleChanOpenAckDuration.Record(ctx.Context(), time.Since(start).Seconds(),
			metric.WithAttributes(attribute.String("contract", contract)),
		)
	}()
	k.Logger(ctx).Debug("HandleChanOpenAck", "port_id", portID, "channel_id", channelID, "counterparty_channel_id", counterpartyChannelID, "counterparty_version", counterpartyVersion)
	icaOwner, err := types.ICAOwnerFromPort(portID)
	if err != nil {
		k.Logger(ctx).Error("HandleChanOpenAck: failed to get ica owner from source port", "error", err)
		return errors.Wrap(err, "failed to get ica owner from port")
	}
	contract = icaOwner.GetContract().String()

	payload, err := keeper.PrepareOpenAckCallbackMessage(contractmanagertypes.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   counterpartyVersion,
	})
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal OpenAckDetails: %v", err)
	}

	_, err = k.sudoKeeper.Sudo(ctx, icaOwner.GetContract(), payload)
	if err != nil {
		k.Logger(ctx).Debug("HandleChanOpenAck: failed to sudo contract on channel open acknowledgement", "error", err)
	}

	return nil
}
