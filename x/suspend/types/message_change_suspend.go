package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgChangeSuspend{}

func NewMsgChangeSuspend(creator string, suspend bool, adminKey string) *MsgChangeSuspend {
	return &MsgChangeSuspend{
		Creator:  creator,
		Suspend:  suspend,
		AdminKey: adminKey,
	}
}

func (msg *MsgChangeSuspend) Route() string {
	return RouterKey
}

func (msg *MsgChangeSuspend) Type() string {
	return "ChangeSuspend"
}

func (msg *MsgChangeSuspend) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgChangeSuspend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgChangeSuspend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
