package types

import (
	"fmt"

	"cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
)

// const interchainAccountIDLimit = 49.
const interchainAccountIDLimit = 128 -
	len("icacontroller-") -
	len("nolus1f6cu6ypvpyh0p8d7pqnps2pduj87hda5t9v4mqrc8ra67xp28uwq4f4ysz") - // just a random contract address
	len(".")

var _ codectypes.UnpackInterfacesMessage = &MsgSubmitTx{}

func (msg *MsgRegisterInterchainAccount) Validate() error {
	if len(msg.ConnectionId) == 0 {
		return ErrEmptyConnectionID
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse FromAddress: %s", msg.FromAddress)
	}

	if len(msg.InterchainAccountId) == 0 {
		return ErrEmptyInterchainAccountID
	}

	if len(msg.InterchainAccountId) > interchainAccountIDLimit {
		return errors.Wrapf(ErrLongInterchainAccountID, "max length is %d, got %d", interchainAccountIDLimit, len(msg.InterchainAccountId))
	}

	return nil
}

func (msg *MsgRegisterInterchainAccount) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

func (msg *MsgRegisterInterchainAccount) Route() string {
	return RouterKey
}

func (msg *MsgRegisterInterchainAccount) Type() string {
	return "register-interchain-account"
}

func (msg *MsgRegisterInterchainAccount) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(msg)
}

//----------------------------------------------------------------

func (msg *MsgSubmitTx) Validate() error {
	if len(msg.ConnectionId) == 0 {
		return ErrEmptyConnectionID
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse FromAddress: %s", msg.FromAddress)
	}

	if len(msg.InterchainAccountId) == 0 {
		return ErrEmptyInterchainAccountID
	}

	if len(msg.Msgs) == 0 {
		return ErrNoMessages
	}

	if msg.Timeout <= 0 {
		return errors.Wrapf(ErrInvalidTimeout, "timeout must be greater than zero")
	}

	return nil
}

func (msg *MsgSubmitTx) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{fromAddress}
}

func (msg *MsgSubmitTx) Route() string {
	return RouterKey
}

func (msg *MsgSubmitTx) Type() string {
	return "submit-tx"
}

func (msg *MsgSubmitTx) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(msg)
}

// PackTxMsgAny marshals the sdk.Msg payload to a protobuf Any type.
func PackTxMsgAny(sdkMsg sdk.Msg) (*codectypes.Any, error) {
	msg, ok := sdkMsg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T", sdkMsg)
	}

	value, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// implements UnpackInterfacesMessage.UnpackInterfaces (https://github.com/cosmos/cosmos-sdk/blob/d07d35f29e0a0824b489c552753e8798710ff5a8/codec/types/interface_registry.go#L60)
func (msg *MsgSubmitTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var sdkMsg sdk.Msg
	for _, m := range msg.Msgs {
		if err := unpacker.UnpackAny(m, &sdkMsg); err != nil {
			return err
		}
	}
	return nil
}

//----------------------------------------------------------------

var _ sdk.Msg = &MsgUpdateParams{}

func (msg *MsgUpdateParams) Route() string {
	return RouterKey
}

func (msg *MsgUpdateParams) Type() string {
	return "update-params"
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{authority}
}

func (msg *MsgUpdateParams) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(msg)
}

func (msg *MsgUpdateParams) Validate() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "authority is invalid")
	}
	return nil
}
