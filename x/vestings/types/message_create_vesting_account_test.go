package types

import (
	"errors"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

// AccAddress returns a sample account address.
func AccAddress() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr)
}

func TestMsgCreateVestingAccount_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateVestingAccount
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateVestingAccount{
				FromAddress: "invalid_address",
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 10)),
				StartTime:   time.Now().Unix(),
				EndTime:     time.Now().Unix() + 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid to address",
			msg: MsgCreateVestingAccount{
				FromAddress: AccAddress().String(),
				ToAddress:   "invalid_address",
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 10)),
				StartTime:   time.Now().Unix(),
				EndTime:     time.Now().Unix() + 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid start time",
			msg: MsgCreateVestingAccount{
				FromAddress: AccAddress().String(),
				ToAddress:   AccAddress().String(),
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 10)),
				StartTime:   0,
				EndTime:     time.Now().Unix() + 1,
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid end time",
			msg: MsgCreateVestingAccount{
				FromAddress: AccAddress().String(),
				ToAddress:   AccAddress().String(),
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 10)),
				StartTime:   time.Now().Unix(),
				EndTime:     0,
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "star time < end time",
			msg: MsgCreateVestingAccount{
				FromAddress: AccAddress().String(),
				ToAddress:   AccAddress().String(),
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 10)),
				StartTime:   time.Now().Unix() - 1,
				EndTime:     time.Now().Unix(),
			},
			err: nil,
		}, {
			name: "star time == end time",
			msg: MsgCreateVestingAccount{
				FromAddress: AccAddress().String(),
				ToAddress:   AccAddress().String(),
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 10)),
				StartTime:   time.Now().Unix(),
				EndTime:     time.Now().Unix(),
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "invalid amount",
			msg: MsgCreateVestingAccount{
				FromAddress: AccAddress().String(),
				ToAddress:   AccAddress().String(),

				StartTime: time.Now().Unix(),
				EndTime:   time.Now().Unix() + 1,
			},
			err: sdkerrors.ErrInvalidCoins,
		}, {
			name: "valid address",
			msg: MsgCreateVestingAccount{
				FromAddress: AccAddress().String(),
				ToAddress:   AccAddress().String(),
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 10)),
				StartTime:   time.Now().Unix(),
				EndTime:     time.Now().Unix() + 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.EqualError(t, errors.Unwrap(err), tt.err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}
