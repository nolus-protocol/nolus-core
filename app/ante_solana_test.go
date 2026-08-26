package app

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

type stubDoneHeight struct {
	height int64
	called bool
}

func (s *stubDoneHeight) GetDoneHeight(_ context.Context, _ string) (int64, error) {
	s.called = true
	return s.height, nil
}

func single(mode txsigning.SignMode) txsigning.SignatureData {
	return &txsigning.SingleSignatureData{SignMode: mode}
}

func TestSignatureUsesSolanaMode(t *testing.T) {
	tests := []struct {
		name string
		data txsigning.SignatureData
		want bool
	}{
		{"offchain", single(txsigning.SignMode_SIGN_MODE_SOLANA_OFFCHAIN), true},
		{"carrier", single(txsigning.SignMode_SIGN_MODE_SOLANA_TX_CARRIER), true},
		{"direct", single(txsigning.SignMode_SIGN_MODE_DIRECT), false},
		{"amino", single(txsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON), false},
		{"nil", nil, false},
		{
			"multi with a nested solana mode",
			&txsigning.MultiSignatureData{Signatures: []txsigning.SignatureData{
				single(txsigning.SignMode_SIGN_MODE_DIRECT),
				single(txsigning.SignMode_SIGN_MODE_SOLANA_OFFCHAIN),
			}},
			true,
		},
		{
			"multi with only standard modes",
			&txsigning.MultiSignatureData{Signatures: []txsigning.SignatureData{
				single(txsigning.SignMode_SIGN_MODE_DIRECT),
				single(txsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON),
			}},
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, signatureUsesSolanaMode(tc.data))
		})
	}
}

func gateTestTx(t *testing.T, testApp *App, mode txsigning.SignMode) sdk.Tx {
	t.Helper()
	builder := testApp.GetTxConfig().NewTxBuilder()
	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	require.NoError(t, builder.SetMsgs(banktypes.NewMsgSend(addr, addr, sdk.NewCoins(sdk.NewInt64Coin("unls", 1)))))
	require.NoError(t, builder.SetSignatures(txsigning.SignatureV2{
		PubKey:   ed25519.GenPrivKey().PubKey(),
		Data:     &txsigning.SingleSignatureData{SignMode: mode},
		Sequence: 0,
	}))
	return builder.GetTx()
}

func TestSolanaSignModeGateDecorator(t *testing.T) {
	testApp, ctx := CreateTestApp(true, t.TempDir())

	newCase := func(mode txsigning.SignMode, doneHeight int64) (*stubDoneHeight, bool, error) {
		keeper := &stubDoneHeight{height: doneHeight}
		decorator := NewSolanaSignModeGateDecorator(keeper, "v0.8.5")
		nextCalled := false
		_, err := decorator.AnteHandle(ctx, gateTestTx(t, testApp, mode), false, func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			nextCalled = true
			return c, nil
		})
		return keeper, nextCalled, err
	}

	t.Run("solana mode rejected before the upgrade is applied", func(t *testing.T) {
		keeper, nextCalled, err := newCase(txsigning.SignMode_SIGN_MODE_SOLANA_OFFCHAIN, 0)
		require.Error(t, err)
		require.False(t, nextCalled, "a gated tx must not reach the rest of the ante chain")
		require.True(t, keeper.called)
	})

	t.Run("solana mode accepted after the upgrade is applied", func(t *testing.T) {
		keeper, nextCalled, err := newCase(txsigning.SignMode_SIGN_MODE_SOLANA_TX_CARRIER, 100)
		require.NoError(t, err)
		require.True(t, nextCalled)
		require.True(t, keeper.called)
	})

	t.Run("standard mode passes through without consulting the upgrade keeper", func(t *testing.T) {
		keeper, nextCalled, err := newCase(txsigning.SignMode_SIGN_MODE_DIRECT, 0)
		require.NoError(t, err)
		require.True(t, nextCalled)
		require.False(t, keeper.called, "non-Solana txs must not pay for a done-height lookup")
	})
}
