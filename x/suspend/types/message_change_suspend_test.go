package types

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgChangeSuspended_ValidateBasic(t *testing.T) {
	hex, _ := sdk.AccAddressFromHex(ed25519.GenPrivKey().PubKey().Address().String())
	tests := []struct {
		name string
		msg  MsgChangeSuspended
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgChangeSuspended{
				FromAddress: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgChangeSuspended{
				FromAddress: hex.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
