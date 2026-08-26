package app

import (
	"testing"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"github.com/stretchr/testify/require"
)

// The pulsar-side enum lives in cosmossdk.io/api, which the fork does not patch,
// so the custom Nolus modes are numeric casts of their wire numbers.
const (
	apiSignModeSolanaOffchain  = signingv1beta1.SignMode(192)
	apiSignModeSolanaTxCarrier = signingv1beta1.SignMode(193)
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
		{"solana offchain", apiSignModeSolanaOffchain},
		{"solana tx carrier", apiSignModeSolanaTxCarrier},
	}

	testApp, _ := CreateTestApp(true, t.TempDir())
	modes := testApp.GetTxConfig().SignModeHandler().SupportedModes()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Contains(t, modes, tc.mode, "the ante handler resolves sign bytes through this map; an unregistered mode is unsignable")
		})
	}
}
