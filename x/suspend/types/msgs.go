package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSuspend{}
var _ sdk.Msg = &MsgUnsuspend{}

func NewMsgSuspend(fromAddress string, suspended bool, blockHeight int64) *MsgSuspend {
	return &MsgSuspend{
		FromAddress: fromAddress,
		BlockHeight: blockHeight,
	}
}

func (msg *MsgSuspend) Route() string {
	return RouterKey
}

func (msg *MsgSuspend) Type() string {
	return "Suspend"
}

func (msg *MsgSuspend) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgSuspend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSuspend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	if msg.BlockHeight < 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "block height must be positive: %d", msg.BlockHeight)
	}
	return nil
}

func NewMsgUnsuspend(fromAddress string) *MsgUnsuspend {
	return &MsgUnsuspend{
		FromAddress: fromAddress,
	}
}

func (msg *MsgUnsuspend) Route() string {
	return RouterKey
}

func (msg *MsgUnsuspend) Type() string {
	return "Unsuspend"
}

func (msg *MsgUnsuspend) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgUnsuspend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnsuspend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
