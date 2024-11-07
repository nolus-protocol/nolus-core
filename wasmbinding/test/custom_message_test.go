package test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/testutil"
	"github.com/Nolus-Protocol/nolus-core/wasmbinding"
	"github.com/Nolus-Protocol/nolus-core/wasmbinding/bindings"
	feetypes "github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
	ictxkeeper "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/keeper"
	ictxtypes "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

const FeeCollectorAddress = "nolus1vguuxez2h5ekltfj9gjd62fs5k4rl2zy5hfrncasykzw08rezpfsd2rhm7"

type CustomMessengerTestSuite struct {
	testutil.IBCConnectionTestSuite
	nolus           *app.App
	ctx             sdk.Context
	messenger       *wasmbinding.CustomMessenger
	contractOwner   sdk.AccAddress
	contractAddress sdk.AccAddress
	contractKeeper  wasmtypes.ContractOpsKeeper
}

func (suite *CustomMessengerTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
	suite.nolus = suite.GetNolusZoneApp(suite.ChainA)
	suite.ctx = suite.ChainA.GetContext()
	suite.messenger = &wasmbinding.CustomMessenger{}
	suite.messenger.Ictxmsgserver = ictxkeeper.NewMsgServerImpl(*suite.nolus.InterchainTxsKeeper)
	suite.messenger.Keeper = *suite.nolus.InterchainTxsKeeper
	suite.messenger.ContractmanagerKeeper = suite.nolus.ContractManagerKeeper
	suite.contractOwner = keeper.RandomAccountAddress(suite.T())

	suite.contractKeeper = keeper.NewDefaultPermissionKeeper(&suite.nolus.WasmKeeper)

	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
	suite.Require().NotEmpty(suite.contractAddress)
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainAccount() {
	// Craft RegisterInterchainAccount message
	msg := bindings.NolusMsg{
		RegisterInterchainAccount: &bindings.RegisterInterchainAccount{
			ConnectionId:        suite.TransferPath.EndpointA.ConnectionID,
			InterchainAccountId: testutil.TestInterchainID,
			RegisterFee:         sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(1_000_000))),
		},
	}

	bankKeeper := suite.nolus.BankKeeper
	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	err := bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(1_000_000))))
	suite.NoError(err)

	// Dispatch RegisterInterchainAccount message
	_, err = suite.executeNolusMsg(suite.contractAddress, msg)
	suite.NoError(err)
}

func (suite *CustomMessengerTestSuite) TestRegisterInterchainAccountLongID() {
	// Store code and instantiate reflect contract
	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
	suite.Require().NotEmpty(suite.contractAddress)

	// Craft RegisterInterchainAccount message
	msg, err := json.Marshal(bindings.NolusMsg{
		RegisterInterchainAccount: &bindings.RegisterInterchainAccount{
			ConnectionId: suite.TransferPath.EndpointA.ConnectionID,
			// the limit is 47, this line is 50 characters long
			InterchainAccountId: "01234567890123456789012345678901234567890123456789",
		},
	})
	suite.NoError(err)

	// Dispatch RegisterInterchainAccount message via DispatchHandler cause we want to catch an error from SDK directly, not from a contract
	_, _, _, err = suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.TransferPath.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
		Custom: msg,
	})
	suite.Error(err)
	suite.ErrorIs(err, ictxtypes.ErrLongInterchainAccountID)
}

func (suite *CustomMessengerTestSuite) TestSubmitTx() {
	// Store code and instantiate reflect contract
	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
	suite.Require().NotEmpty(suite.contractAddress)

	senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(int64(10_000_000))))
	bankKeeper := suite.nolus.BankKeeper
	err := bankKeeper.SendCoins(suite.ctx, senderAddress, suite.contractAddress, coinsAmnt)
	suite.NoError(err)

	icaowner := suite.contractAddress.String() + ".0"

	path := testutil.NewICAPath(suite.ChainA, suite.ChainB, icaowner)
	suite.Coordinator.SetupConnections(path)

	err = suite.SetupICAPath(path, icaowner)
	suite.Require().NoError(err)

	events, data, _, err := suite.messenger.DispatchMsg(
		suite.ctx,
		suite.contractAddress,
		path.EndpointA.ChannelConfig.PortID,
		types.CosmosMsg{
			Custom: suite.craftMarshaledMsgSubmitTxWithNumMsgs(1, path),
		},
	)
	suite.NoError(err)

	var response ictxtypes.MsgSubmitTxResponse
	err = json.Unmarshal(data[0], &response)
	suite.NoError(err)
	suite.Nil(events)
	suite.Equal(uint64(1), response.SequenceId)
	suite.Equal("channel-1", response.Channel)
}

func (suite *CustomMessengerTestSuite) TestSubmitTxTooMuchTxs() {
	// Store code and instantiate reflect contract
	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
	suite.contractAddress = suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
	suite.Require().NotEmpty(suite.contractAddress)

	icaowner := suite.contractAddress.String() + ".0"

	path := testutil.NewICAPath(suite.ChainA, suite.ChainB, icaowner)
	suite.Coordinator.SetupConnections(path)

	err := suite.SetupICAPath(path, icaowner)
	suite.Require().NoError(err)

	_, _, _, err = suite.messenger.DispatchMsg(
		suite.ctx,
		suite.contractAddress,
		path.EndpointA.ChannelConfig.PortID,
		types.CosmosMsg{
			Custom: suite.craftMarshaledMsgSubmitTxWithNumMsgs(20, path),
		},
	)
	suite.ErrorContains(err, "MsgSubmitTx contains more messages than allowed")
}

// func (suite *CustomMessengerTestSuite) TestResubmitFailureAck() {
// 	// Add failure
// 	packet := ibcchanneltypes.Packet{}
// 	ack := ibcchanneltypes.Acknowledgement{
// 		Response: &ibcchanneltypes.Acknowledgement_Result{Result: []byte("Result")},
// 	}
// 	payload, err := contractmanagerkeeper.PrepareSudoCallbackMessage(packet, &ack)
// 	require.NoError(suite.T(), err)
// 	failureID := suite.messenger.ContractmanagerKeeper.GetNextFailureIDKey(suite.ctx, suite.contractAddress.String())
// 	suite.messenger.ContractmanagerKeeper.AddContractFailure(suite.ctx, suite.contractAddress.String(), payload, "test error")

// 	// Craft message
// 	msg := bindings.NolusMsg{
// 		ResubmitFailure: &bindings.ResubmitFailure{
// 			FailureId: failureID,
// 		},
// 	}

// 	// Dispatch
// 	data, err := suite.executeNolusMsg(suite.contractAddress, msg)
// 	suite.NoError(err)

// 	var expected contractmanagertypes.Failure
// 	err = expected.Unmarshal(data)
// 	suite.NoError(err)
// 	suite.Equal(expected.Id, failureID)
// }

// func (suite *CustomMessengerTestSuite) TestResubmitFailureTimeout() {
// 	// Store code and instantiate reflect contract
// 	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
// 	suite.contractAddress = suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
// 	suite.Require().NotEmpty(suite.contractAddress)

// 	// Add failure
// 	packet := ibcchanneltypes.Packet{}
// 	payload, err := contractmanagerkeeper.PrepareSudoCallbackMessage(packet, nil)
// 	require.NoError(suite.T(), err)
// 	failureID := suite.messenger.ContractmanagerKeeper.GetNextFailureIDKey(suite.ctx, suite.contractAddress.String())
// 	suite.messenger.ContractmanagerKeeper.AddContractFailure(suite.ctx, suite.contractAddress.String(), payload, "test error")

// 	// Craft message
// 	msg, err := json.Marshal(bindings.NolusMsg{
// 		ResubmitFailure: &bindings.ResubmitFailure{
// 			FailureId: failureID,
// 		},
// 	})
// 	suite.NoError(err)

// 	icaowner := suite.contractAddress.String() + ".0"

// 	path := testutil.NewICAPath(suite.ChainA, suite.ChainB, icaowner)
// 	suite.Coordinator.SetupConnections(path)

// 	err = suite.SetupICAPath(path, icaowner)
// 	suite.Require().NoError(err)

// 	// Dispatch
// 	events, data, _, err := suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
// 		Custom: msg,
// 	})
// 	suite.NoError(err)
// 	suite.Nil(events)
// 	expected, err := json.Marshal(&bindings.ResubmitFailureResponse{FailureId: failureID})
// 	suite.NoError(err)
// 	suite.Equal([][]uint8{expected}, data)
// }

// func (suite *CustomMessengerTestSuite) TestResubmitFailureFromDifferentContract() {
// 	// Store code and instantiate reflect contract
// 	codeID := suite.StoreTestCode(suite.ctx, suite.contractOwner, "../testdata/reflect.wasm")
// 	suite.contractAddress = suite.InstantiateTestContract(suite.ctx, suite.contractOwner, codeID)
// 	suite.Require().NotEmpty(suite.contractAddress)

// 	// Add failure
// 	packet := ibcchanneltypes.Packet{}
// 	ack := ibcchanneltypes.Acknowledgement{
// 		Response: &ibcchanneltypes.Acknowledgement_Error{Error: "ErrorSudoPayload"},
// 	}
// 	failureID := suite.messenger.ContractmanagerKeeper.GetNextFailureIDKey(suite.ctx, testutil.TestOwnerAddress)
// 	payload, err := contractmanagerkeeper.PrepareSudoCallbackMessage(packet, &ack)
// 	require.NoError(suite.T(), err)
// 	suite.messenger.ContractmanagerKeeper.AddContractFailure(suite.ctx, testutil.TestOwnerAddress, payload, "test error")

// 	// Craft message
// 	msg, err := json.Marshal(bindings.NolusMsg{
// 		ResubmitFailure: &bindings.ResubmitFailure{
// 			FailureId: failureID,
// 		},
// 	})
// 	suite.NoError(err)

// 	// Dispatch
// 	_, _, _, err = suite.messenger.DispatchMsg(suite.ctx, suite.contractAddress, suite.Path.EndpointA.ChannelConfig.PortID, types.CosmosMsg{
// 		Custom: msg,
// 	})
// 	suite.ErrorContains(err, "no failure found to resubmit: not found")
// }

func (suite *CustomMessengerTestSuite) craftMarshaledMsgSubmitTxWithNumMsgs(numMsgs int, path *ibctesting.Path) (result []byte) {
	msg := bindings.ProtobufAny{
		TypeURL: "/cosmos.staking.v1beta1.MsgDelegate",
		Value:   []byte{26, 10, 10, 5, 115, 116, 97, 107, 101, 18, 1, 48},
	}
	msgs := make([]bindings.ProtobufAny, 0, numMsgs)
	for i := 0; i < numMsgs; i++ {
		msgs = append(msgs, msg)
	}
	result, err := json.Marshal(struct {
		SubmitTx bindings.SubmitTx `json:"submit_tx"`
	}{
		SubmitTx: bindings.SubmitTx{
			ConnectionId:        path.EndpointA.ConnectionID,
			InterchainAccountId: "0",
			Msgs:                msgs,
			Memo:                "Jimmy",
			Timeout:             2000,
			Fee: feetypes.Fee{
				RecvFee:    sdk.NewCoins(),
				AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(1000))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(1000))),
			},
		},
	})
	suite.NoError(err)
	return
}

func (suite *CustomMessengerTestSuite) executeCustomMsg(contractAddress sdk.AccAddress, fullMsg json.RawMessage) (data []byte, err error) {
	customMsg := types.CosmosMsg{
		Custom: fullMsg,
	}

	type ExecuteMsg struct {
		ReflectMsg struct {
			Msgs []types.CosmosMsg `json:"msgs"`
		} `json:"reflect_msg"`
	}

	execMsg := ExecuteMsg{ReflectMsg: struct {
		Msgs []types.CosmosMsg `json:"msgs"`
	}(struct{ Msgs []types.CosmosMsg }{Msgs: []types.CosmosMsg{customMsg}})}

	msg, err := json.Marshal(execMsg)
	suite.NoError(err)

	data, err = suite.contractKeeper.Execute(suite.ctx, contractAddress, suite.contractOwner, msg, nil)

	return
}

func (suite *CustomMessengerTestSuite) executeNolusMsg(contractAddress sdk.AccAddress, fullMsg bindings.NolusMsg) (data []byte, err error) {
	fullMsgBz, err := json.Marshal(fullMsg)
	suite.NoError(err)

	return suite.executeCustomMsg(contractAddress, fullMsgBz)
}

func TestMessengerTestSuite(t *testing.T) {
	suite.Run(t, new(CustomMessengerTestSuite))
}
