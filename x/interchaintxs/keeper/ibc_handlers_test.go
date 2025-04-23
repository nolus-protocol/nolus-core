package keeper_test

import (
	"fmt"
	"testing"

	types2 "cosmossdk.io/store/types"

	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/testutil"
	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/interchaintxs/keeper"
	mock_types "github.com/Nolus-Protocol/nolus-core/testutil/mocks/interchaintxs/types"
	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
)

const ICAId = ".ica0"

func TestHandleAcknowledgement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	icak, infCtx := testkeeper.InterchainTxsKeeper(t, wmKeeper, feeKeeper, nil, nil, nil, bankKeeper, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})
	ctx := infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))

	resACK := channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{Result: []byte("Result")},
	}
	resAckData, err := channeltypes.SubModuleCdc.MarshalJSON(&resACK)
	require.NoError(t, err)
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.ControllerPortPrefix + testutil.TestOwnerAddress + ICAId,
		SourceChannel: "channel-0",
	}
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	relayerBech32 := "nolus1f6cu6ypvpyh0p8d7pqnps2pduj87hda5t9v4mqrc8ra67xp28uwq4f4ysz"
	relayerAddress := sdk.MustAccAddressFromBech32(relayerBech32)

	err = icak.HandleAcknowledgement(ctx, channeltypes.Packet{}, nil, relayerAddress)
	require.ErrorContains(t, err, "failed to get ica owner from port")

	err = icak.HandleAcknowledgement(ctx, p, nil, relayerAddress)
	require.ErrorContains(t, err, "cannot unmarshal ICS-27 packet acknowledgement")

	msgAck, err := keeper.PrepareSudoCallbackMessage(p, &resACK)
	require.NoError(t, err)

	// success contract SudoResponse
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msgAck)
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)

	// error contract SudoResponse
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msgAck).Return(nil, fmt.Errorf("error sudoResponse"))
	err = icak.HandleAcknowledgement(ctx, p, resAckData, relayerAddress)
	require.NoError(t, err)
}

func TestHandleTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	feeKeeper := mock_types.NewMockFeeRefunderKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	icak, infCtx := testkeeper.InterchainTxsKeeper(t, wmKeeper, feeKeeper, nil, nil, nil, bankKeeper, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})
	ctx := infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	relayerBech32 := "nolus1f6cu6ypvpyh0p8d7pqnps2pduj87hda5t9v4mqrc8ra67xp28uwq4f4ysz"
	relayerAddress := sdk.MustAccAddressFromBech32(relayerBech32)
	p := channeltypes.Packet{
		Sequence:      100,
		SourcePort:    icatypes.ControllerPortPrefix + testutil.TestOwnerAddress + ICAId,
		SourceChannel: "channel-0",
	}

	msgAck, err := keeper.PrepareSudoCallbackMessage(p, nil)
	require.NoError(t, err)

	err = icak.HandleTimeout(ctx, channeltypes.Packet{}, relayerAddress)
	require.ErrorContains(t, err, "failed to get ica owner from port")

	// contract success
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msgAck)
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)

	// contract error
	ctx = infCtx.WithGasMeter(types2.NewGasMeter(1_000_000_000_000))
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msgAck).Return(nil, fmt.Errorf("SudoTimeout error"))
	err = icak.HandleTimeout(ctx, p, relayerAddress)
	require.NoError(t, err)
}

func TestHandleChanOpenAck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wmKeeper := mock_types.NewMockWasmKeeper(ctrl)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)
	icak, ctx := testkeeper.InterchainTxsKeeper(t, wmKeeper, nil, nil, nil, nil, bankKeeper, func(_ sdk.Context) string {
		return TestFeeCollectorAddr
	})
	portID := icatypes.ControllerPortPrefix + testutil.TestOwnerAddress + ICAId
	contractAddress := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)
	const channelID = "channel-0"
	counterpartyChannelID := "channel-1"

	err := icak.HandleChanOpenAck(ctx, "", channelID, counterpartyChannelID, "1")
	require.ErrorContains(t, err, "failed to get ica owner from port")

	msg, err := keeper.PrepareOpenAckCallbackMessage(types.OpenAckDetails{
		PortID:                portID,
		ChannelID:             channelID,
		CounterpartyChannelID: counterpartyChannelID,
		CounterpartyVersion:   "1",
	})
	require.NoError(t, err)

	// sudo error
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msg).Return(nil, fmt.Errorf("SudoOnChanOpenAck error"))
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	require.NoError(t, err)

	// sudo success
	wmKeeper.EXPECT().Sudo(ctx, contractAddress, msg)
	err = icak.HandleChanOpenAck(ctx, portID, channelID, counterpartyChannelID, "1")
	require.NoError(t, err)
}
