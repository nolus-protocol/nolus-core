package app

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdked25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// A nil FeePayerSource fails closed with a distinct error, so a refactor that
// dropped the keeper from CustomSignModeHandlers in app.go would break no other
// test while bricking the sign mode on chain. This drives the handler the app's
// TxConfig registered with a carrier whose fee payer is not the parameter: only
// a handler that read the keeper's value can reach that rejection.
func TestSolanaCarrierSignModeReadsTheFeePayerFromTheKeeper(t *testing.T) {
	testApp, ctx := CreateTestApp(true, t.TempDir())
	require.NoError(t, testApp.SolanaCarrierKeeper.SetParams(ctx, solanacarrier.DefaultParams()))

	signerKey := make([]byte, 32)
	signerKey[0] = 0x01
	wrongFeePayer := make([]byte, 32)
	wrongFeePayer[0] = 0x02

	// Legacy message: header (2 required signatures, 0 read-only signed, 0
	// read-only unsigned), two account keys, a blockhash, no instructions.
	message := []byte{0x02, 0x00, 0x00, 0x02}
	message = append(message, wrongFeePayer...)
	message = append(message, signerKey...)
	message = append(message, make([]byte, 32)...)
	message = append(message, 0x00)

	carrierAny, err := codectypes.NewAnyWithValue(&solanacarrier.SolanaCarrier{Message: message})
	require.NoError(t, err)

	builder, ok := testApp.GetTxConfig().NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	require.True(t, ok)
	require.NoError(t, builder.SetMsgs(&banktypes.MsgSend{
		FromAddress: sdk.AccAddress(make([]byte, 20)).String(),
		ToAddress:   sdk.AccAddress(make([]byte, 20)).String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 1)),
	}))
	builder.SetNonCriticalExtensionOptions(carrierAny)
	adaptable, ok := builder.GetTx().(authsigning.V2AdaptableTx)
	require.True(t, ok)

	pubKeyAny, err := codectypes.NewAnyWithValue(&sdked25519.PubKey{Key: signerKey})
	require.NoError(t, err)
	signerData := txsigning.SignerData{
		ChainID: "nolus-1",
		Address: sdk.AccAddress(make([]byte, 20)).String(),
		PubKey:  &anypb.Any{TypeUrl: pubKeyAny.TypeUrl, Value: pubKeyAny.Value},
	}

	_, err = testApp.GetTxConfig().SignModeHandler().GetSignBytes(
		ctx,
		signingv1beta1.SignMode_SIGN_MODE_SOLANA_TX_CARRIER,
		signerData,
		adaptable.GetSigningTxData(),
	)
	require.ErrorContains(t, err, "requires the configured fee payer at account index 0")
}
