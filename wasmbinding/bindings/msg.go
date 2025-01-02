//nolint:revive,stylecheck  // if we change the names of var-naming things here, we harm some kind of mapping.
package bindings

import (
	paramChange "cosmossdk.io/x/params/types/proposal"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feetypes "github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
	transferwrappertypes "github.com/Nolus-Protocol/nolus-core/x/transfer/types"
)

// ProtobufAny is a hack-struct to serialize protobuf Any message into JSON object.
type ProtobufAny struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

// NolusMsg is used like a sum type to hold one of custom Nolus messages.
// Follow https://github.com/neutron-org/neutron/neutron-contracts/tree/main/packages/bindings/src/msg.rs
// for more information.
type NolusMsg struct {
	SubmitTx                  *SubmitTx                         `json:"submit_tx,omitempty"`
	RegisterInterchainAccount *RegisterInterchainAccount        `json:"register_interchain_account,omitempty"`
	IBCTransfer               *transferwrappertypes.MsgTransfer `json:"ibc_transfer,omitempty"`

	// Contractmanager types
	/// A contract that has failed acknowledgement can resubmit it
	ResubmitFailure *ResubmitFailure `json:"resubmit_failure,omitempty"`
}

// SubmitTx submits interchain transaction on a remote chain.
type SubmitTx struct {
	ConnectionId        string        `json:"connection_id"`
	InterchainAccountId string        `json:"interchain_account_id"`
	Msgs                []ProtobufAny `json:"msgs"`
	Memo                string        `json:"memo"`
	Timeout             uint64        `json:"timeout"`
	Fee                 feetypes.Fee  `json:"fee"`
}

// RegisterInterchainAccount creates account on remote chain.
type RegisterInterchainAccount struct {
	ConnectionId        string    `json:"connection_id"`
	InterchainAccountId string    `json:"interchain_account_id"`
	RegisterFee         sdk.Coins `json:"register_fee,omitempty"`
}

// RegisterInterchainAccountResponse holds response for RegisterInterchainAccount.
type RegisterInterchainAccountResponse struct {
	ChannelId string `json:"channel_id"`
	PortId    string `json:"port_id"`
}

type ParamChangeProposal struct {
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	ParamChanges []paramChange.ParamChange `json:"param_changes"`
}

type SoftwareUpgradeProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Plan        Plan   `json:"plan"`
}

type CancelSoftwareUpgradeProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Plan struct {
	Name   string `json:"name"`
	Height int64  `json:"height"`
	Info   string `json:"info"`
}

// MsgExecuteContract defined separate from wasmtypes since we can get away with just passing the string into bindings.
type MsgExecuteContract struct {
	// Contract is the address of the smart contract
	Contract string `json:"contract,omitempty"`
	// Msg json encoded message to be passed to the contract
	Msg string `json:"msg,omitempty"`
}

type ResubmitFailure struct {
	FailureId uint64 `json:"failure_id"`
}

type ResubmitFailureResponse struct {
	FailureId uint64 `json:"failure_id"`
}
