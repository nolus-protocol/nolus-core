package wasmbinding

import (
	"encoding/json"

	"cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Nolus-Protocol/nolus-core/wasmbinding/bindings"

	contractmanagerkeeper "github.com/Nolus-Protocol/nolus-core/x/contractmanager/keeper"
	ictxkeeper "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/keeper"
	ictxtypes "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
	transferwrapperkeeper "github.com/Nolus-Protocol/nolus-core/x/transfer/keeper"
	transferwrappertypes "github.com/Nolus-Protocol/nolus-core/x/transfer/types"
)

func CustomMessageDecorator(
	ictx *ictxkeeper.Keeper,
	transferKeeper transferwrapperkeeper.KeeperTransferWrapper,
	contractmanagerKeeper *contractmanagerkeeper.Keeper,
) func(messenger wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			Keeper:                *ictx,
			Wrapped:               old,
			Ictxmsgserver:         ictxkeeper.NewMsgServerImpl(*ictx),
			transferKeeper:        transferKeeper,
			ContractmanagerKeeper: contractmanagerKeeper,
		}
	}
}

type CustomMessenger struct {
	Keeper                ictxkeeper.Keeper
	Wrapped               wasmkeeper.Messenger
	Ictxmsgserver         ictxtypes.MsgServer
	transferKeeper        transferwrapperkeeper.KeeperTransferWrapper
	ContractmanagerKeeper *contractmanagerkeeper.Keeper
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	if msg.Custom != nil {
		var contractMsg bindings.NolusMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			ctx.Logger().Debug("json.Unmarshal: failed to decode incoming custom cosmos message",
				"from_address", contractAddr.String(),
				"message", string(msg.Custom),
				"error", err,
			)
			return nil, nil, nil, errors.Wrap(err, "failed to decode incoming custom cosmos message")
		}

		if contractMsg.SubmitTx != nil {
			return m.submitTx(ctx, contractAddr, contractMsg.SubmitTx)
		}
		if contractMsg.RegisterInterchainAccount != nil {
			return m.registerInterchainAccount(ctx, contractAddr, contractMsg.RegisterInterchainAccount)
		}
		if contractMsg.IBCTransfer != nil {
			return m.ibcTransfer(ctx, contractAddr, *contractMsg.IBCTransfer)
		}
		if contractMsg.ResubmitFailure != nil {
			return m.resubmitFailure(ctx, contractAddr, contractMsg.ResubmitFailure)
		}
	}

	return m.Wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func (m *CustomMessenger) ibcTransfer(ctx sdk.Context, contractAddr sdk.AccAddress, ibcTransferMsg transferwrappertypes.MsgTransfer) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	ibcTransferMsg.Sender = contractAddr.String()

	response, err := m.transferKeeper.Transfer(ctx, &ibcTransferMsg)
	if err != nil {
		ctx.Logger().Debug("transferServer.Transfer: failed to transfer",
			"from_address", contractAddr.String(),
			"msg", ibcTransferMsg,
			"error", err,
		)
		return nil, nil, nil, errors.Wrap(err, "failed to execute IBCTransfer")
	}

	data, err := json.Marshal(response)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal MsgTransferResponse response to JSON",
			"from_address", contractAddr.String(),
			"msg", response,
			"error", err,
		)
		return nil, nil, nil, errors.Wrap(err, "marshal json failed")
	}

	ctx.Logger().Debug("ibcTransferMsg completed",
		"from_address", contractAddr.String(),
		"msg", ibcTransferMsg,
	)

	anyResp, err := types.NewAnyWithValue(response)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to convert {%T} to Any", response)
	}
	msgResponses := [][]*types.Any{{anyResp}}
	return nil, [][]byte{data}, msgResponses, nil
}

func (m *CustomMessenger) submitTx(ctx sdk.Context, contractAddr sdk.AccAddress, submitTx *bindings.SubmitTx) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	response, err := m.performSubmitTx(ctx, contractAddr, submitTx)
	if err != nil {
		ctx.Logger().Debug("performSubmitTx: failed to submit interchain transaction",
			"from_address", contractAddr.String(),
			"connection_id", submitTx.ConnectionId,
			"interchain_account_id", submitTx.InterchainAccountId,
			"error", err,
		)
		return nil, nil, nil, errors.Wrap(err, "failed to submit interchain transaction")
	}

	data, err := json.Marshal(response)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal submitTx response to JSON",
			"from_address", contractAddr.String(),
			"connection_id", submitTx.ConnectionId,
			"interchain_account_id", submitTx.InterchainAccountId,
			"error", err,
		)
		return nil, nil, nil, errors.Wrap(err, "marshal json failed")
	}

	ctx.Logger().Debug("interchain transaction submitted",
		"from_address", contractAddr.String(),
		"connection_id", submitTx.ConnectionId,
		"interchain_account_id", submitTx.InterchainAccountId,
	)

	anyResp, err := types.NewAnyWithValue(response)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to convert {%T} to Any", response)
	}
	msgResponses := [][]*types.Any{{anyResp}}
	return nil, [][]byte{data}, msgResponses, nil
}

func (m *CustomMessenger) performSubmitTx(ctx sdk.Context, contractAddr sdk.AccAddress, submitTx *bindings.SubmitTx) (*ictxtypes.MsgSubmitTxResponse, error) {
	tx := ictxtypes.MsgSubmitTx{
		FromAddress:         contractAddr.String(),
		ConnectionId:        submitTx.ConnectionId,
		Memo:                submitTx.Memo,
		InterchainAccountId: submitTx.InterchainAccountId,
		Timeout:             submitTx.Timeout,
		Fee:                 submitTx.Fee,
	}
	for _, msg := range submitTx.Msgs {
		tx.Msgs = append(tx.Msgs, &types.Any{
			TypeUrl: msg.TypeURL,
			Value:   msg.Value,
		})
	}

	response, err := m.Ictxmsgserver.SubmitTx(ctx, &tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to submit interchain transaction")
	}

	return response, nil
}

func (m *CustomMessenger) registerInterchainAccount(ctx sdk.Context, contractAddr sdk.AccAddress, reg *bindings.RegisterInterchainAccount) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	response, err := m.performRegisterInterchainAccount(ctx, contractAddr, reg)
	if err != nil {
		ctx.Logger().Debug("performRegisterInterchainAccount: failed to register interchain account",
			"from_address", contractAddr.String(),
			"connection_id", reg.ConnectionId,
			"interchain_account_id", reg.InterchainAccountId,
			"error", err,
		)
		return nil, nil, nil, errors.Wrap(err, "failed to register interchain account")
	}

	data, err := json.Marshal(response)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal register interchain account response to JSON",
			"from_address", contractAddr.String(),
			"connection_id", reg.ConnectionId,
			"interchain_account_id", reg.InterchainAccountId,
			"error", err,
		)
		return nil, nil, nil, errors.Wrap(err, "marshal json failed")
	}

	ctx.Logger().Debug("registered interchain account",
		"from_address", contractAddr.String(),
		"connection_id", reg.ConnectionId,
		"interchain_account_id", reg.InterchainAccountId,
	)

	anyResp, err := types.NewAnyWithValue(response)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to convert {%T} to Any", response)
	}
	msgResponses := [][]*types.Any{{anyResp}}
	return nil, [][]byte{data}, msgResponses, nil
}

func (m *CustomMessenger) performRegisterInterchainAccount(ctx sdk.Context, contractAddr sdk.AccAddress, reg *bindings.RegisterInterchainAccount) (*ictxtypes.MsgRegisterInterchainAccountResponse, error) {
	msg := ictxtypes.MsgRegisterInterchainAccount{
		FromAddress:         contractAddr.String(),
		ConnectionId:        reg.ConnectionId,
		InterchainAccountId: reg.InterchainAccountId,
		RegisterFee:         getRegisterFee(reg.RegisterFee),
	}

	response, err := m.Ictxmsgserver.RegisterInterchainAccount(ctx, &msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register interchain account")
	}

	return response, nil
}

func (m *CustomMessenger) resubmitFailure(ctx sdk.Context, contractAddr sdk.AccAddress, resubmitFailure *bindings.ResubmitFailure) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	failure, err := m.ContractmanagerKeeper.GetFailure(ctx, contractAddr, resubmitFailure.FailureId)
	if err != nil {
		return nil, nil, nil, errors.Wrap(sdkerrors.ErrNotFound, "no failure found to resubmit")
	}

	err = m.ContractmanagerKeeper.ResubmitFailure(ctx, contractAddr, failure)

	if err != nil {
		ctx.Logger().Error("failed to resubmitFailure",
			"from_address", contractAddr.String(),
			"error", err,
		)
		return nil, nil, nil, errors.Wrap(err, "failed to resubmitFailure")
	}

	resp := bindings.ResubmitFailureResponse{FailureId: failure.Id}
	data, err := json.Marshal(&resp)
	if err != nil {
		ctx.Logger().Error("json.Marshal: failed to marshal remove resubmitFailure response to JSON",
			"from_address", contractAddr.String(),
			"error", err,
		)
		return nil, nil, nil, errors.Wrap(err, "marshal json failed")
	}

	anyResp, err := types.NewAnyWithValue(failure)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to convert {%T} to Any", failure)
	}
	msgResponses := [][]*types.Any{{anyResp}}
	return nil, [][]byte{data}, msgResponses, nil
}

func getRegisterFee(fee sdk.Coins) sdk.Coins {
	if fee == nil {
		return make(sdk.Coins, 0)
	}
	return fee
}
