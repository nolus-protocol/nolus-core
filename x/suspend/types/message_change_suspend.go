package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgChangeSuspended{}

func NewMsgChangeSuspended(fromAddress string, suspended bool, blockHeight int64) *MsgChangeSuspended {
	return &MsgChangeSuspended{
		FromAddress: fromAddress,
		Suspended:     suspended,
		BlockHeight: blockHeight,
	}
}

func (msg *MsgChangeSuspended) Route() string {
	return RouterKey
}

func (msg *MsgChangeSuspended) Type() string {
	return "ChangeSuspend"
}

func (msg *MsgChangeSuspended) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgChangeSuspended) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgChangeSuspended) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	if msg.BlockHeight < 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "block height must be positive: %d", msg.BlockHeight)
	}
	return nil
}
