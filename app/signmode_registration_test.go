package app

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestTxConfigRegistersEverySignMode(t *testing.T) {
	tests := []struct {
		name string
		mode signingv1beta1.SignMode
	}{
		{"direct", signingv1beta1.SignMode_SIGN_MODE_DIRECT},
		{"direct aux", signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX},
		{"legacy amino json", signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON},
		{"textual", signingv1beta1.SignMode_SIGN_MODE_TEXTUAL},
		{"eip 191", signingv1beta1.SignMode_SIGN_MODE_EIP_191}, //nolint:staticcheck
		{"solana offchain", signingv1beta1.SignMode_SIGN_MODE_SOLANA_OFFCHAIN},
		{"solana tx carrier", signingv1beta1.SignMode_SIGN_MODE_SOLANA_TX_CARRIER},
	}

	testApp, _ := CreateTestApp(true, t.TempDir())
	modes := testApp.GetTxConfig().SignModeHandler().SupportedModes()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Contains(t, modes, tc.mode, "the ante handler resolves sign bytes through this map; an unregistered mode is unsignable")
		})
	}
}

func TestSolanaSignModeWireNumbers(t *testing.T) {
	tests := []struct {
		name string
		mode signingv1beta1.SignMode
		want int32
	}{
		{"solana offchain", signingv1beta1.SignMode_SIGN_MODE_SOLANA_OFFCHAIN, 192},
		{"solana tx carrier", signingv1beta1.SignMode_SIGN_MODE_SOLANA_TX_CARRIER, 193},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, int32(tc.mode), "wallets hardcode this wire number; it must never move")
		})
	}
}

// The gogo enum (github.com/cosmos/cosmos-sdk) and the pulsar enum (cosmossdk.io/api)
// come from separately tagged fork modules; the hybrid resolver silently prefers the
// pulsar copy on mismatch, so a mismatched pin surfaces nowhere else.
func TestSignModeDescriptorsAgree(t *testing.T) {
	gz, path := txsigning.SignMode(0).EnumDescriptor()
	zr, err := gzip.NewReader(bytes.NewReader(gz))
	require.NoError(t, err)
	raw, err := io.ReadAll(zr)
	require.NoError(t, err)
	var fd descriptorpb.FileDescriptorProto
	require.NoError(t, proto.Unmarshal(raw, &fd))
	gogoValues := fd.EnumType[path[0]].GetValue()

	pulsarValues := signingv1beta1.SignMode(0).Descriptor().Values()
	require.Equal(t, pulsarValues.Len(), len(gogoValues), "gogo and pulsar SignMode descriptors must come from the same proto")
	for i, gogoValue := range gogoValues {
		pulsarValue := pulsarValues.Get(i)
		require.Equal(t, string(pulsarValue.Name()), gogoValue.GetName())
		require.Equal(t, int32(pulsarValue.Number()), gogoValue.GetNumber())
	}
	require.NotNil(t, pulsarValues.ByNumber(192), "SIGN_MODE_SOLANA_OFFCHAIN missing from the descriptor")
	require.NotNil(t, pulsarValues.ByNumber(193), "SIGN_MODE_SOLANA_TX_CARRIER missing from the descriptor")
}
