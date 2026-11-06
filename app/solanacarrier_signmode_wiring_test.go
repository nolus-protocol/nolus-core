package app

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdked25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// A nil ParamsSource fails closed with a distinct error, so a refactor that
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

const solflareLighthouseTxBase64 = "CsQGCpoBCiMvY29zbW9zLnN0YWtpbmcudjFiZXRhMS5Nc2dEZWxlZ2F0ZRJzCixub2x1czFyMDZwcXp1eDZmODVkdG53NWpkZDluYTRkMHRodHBuZXo0MzRydBIzbm9sdXN2YWxvcGVyMXI4dGVsbjN5M2dweXRmZnNzM3hxbHpxa3lzOGRwd2hlejh4MnI4Gg4KBHVubHMSBjEwMDAwMPp/owUKJS9ub2x1cy5zb2xhbmFjYXJyaWVyLnYxLlNvbGFuYUNhcnJpZXIS+QQK9gQCAQMFzvwDfBFDbchbaUO0PJTsUMBJNpbLfnl+hMF2MDjprXAJO9bpriBNWe3ovRG2JqFxvMpPJsSsDULO/qE8hZ3/pQMGRm/lIRcy/+ytunLDm+e8jOW7xfcSayxDmzpAAAAABN+teWL/sd2SXQqftebQDOYZW6i7OpH9B++YYMXpe7gFSlNamSkhBk0k6HFg2jh8fDW13bySu4HkH6hAQQVEjaGylkBHD9Mp81zOJ8hlVS/5c+TskrJI3edRpcJiFoQ4BAIABQKIVgMABAEB+gJ7ImFjY291bnRfbnVtYmVyIjoiNjczODQiLCJjaGFpbl9pZCI6InBpcmluLTEiLCJmZWUiOnsiYW1vdW50IjpbeyJhbW91bnQiOiIyMDMzNiIsImRlbm9tIjoidW5scyJ9XSwiZ2FzIjoiODEzNDM5In0sIm1lbW8iOiIiLCJtc2dzIjpbeyJ0eXBlIjoiY29zbW9zLXNkay9Nc2dEZWxlZ2F0ZSIsInZhbHVlIjp7ImFtb3VudCI6eyJhbW91bnQiOiIxMDAwMDAiLCJkZW5vbSI6InVubHMifSwiZGVsZWdhdG9yX2FkZHJlc3MiOiJub2x1czFyMDZwcXp1eDZmODVkdG53NWpkZDluYTRkMHRodHBuZXo0MzRydCIsInZhbGlkYXRvcl9hZGRyZXNzIjoibm9sdXN2YWxvcGVyMXI4dGVsbjN5M2dweXRmZnNzM3hxbHpxa3lzOGRwd2hlejh4MnI4In19XSwic2VxdWVuY2UiOiI1In0CAAkDoIYBAAAAAAADAQEaBgUDAAAAAAAAAAAABAMAAAEAAAAAAAAAAAASZQpOCkMKHS9jb3Ntb3MuY3J5cHRvLmVkMjU1MTkuUHViS2V5EiIKIAk71umuIE1Z7ei9EbYmoXG8yk8mxKwNQs7+oTyFnf+lEgUKAwjBARgFEhMKDQoEdW5scxIFMjAzMzYQ/9IxGkBWv1Clx5hPh0hLSjbfDdYeFKmRyoizdJjsYWoJcSwXR73G91Qj+Cdnx0QUlU1Bb1ZjX/6maZ+MC6mO1L1djd4M"

const (
	solflareChainID       = "pirin-1" // from the carrier memo's amino JSON
	solflareAccountNumber = 67384     // from the carrier memo's amino JSON
)

func TestSolanaCarrierVerifiesTheSolflareLighthouseCapture(t *testing.T) {
	testApp, ctx := CreateTestApp(true, t.TempDir())
	txConfig := testApp.GetTxConfig()

	// A pirin-1 MsgDelegate signed with Solflare on a Ledger; Solflare added a
	// Lighthouse instruction to the carrier.
	txBytes, err := base64.StdEncoding.DecodeString(solflareLighthouseTxBase64)
	require.NoError(t, err)
	tx, err := txConfig.TxDecoder()(txBytes)
	require.NoError(t, err)

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	require.True(t, ok)
	pubKeys, err := sigTx.GetPubKeys()
	require.NoError(t, err)
	require.Len(t, pubKeys, 1)
	signatures, err := sigTx.GetSignaturesV2()
	require.NoError(t, err)
	require.Len(t, signatures, 1)
	single, ok := signatures[0].Data.(*sdksigning.SingleSignatureData)
	require.True(t, ok)
	require.Equal(t, sdksigning.SignMode_SIGN_MODE_SOLANA_TX_CARRIER, single.SignMode)

	pubKeyAny, err := codectypes.NewAnyWithValue(pubKeys[0])
	require.NoError(t, err)
	signerData := txsigning.SignerData{
		ChainID:       solflareChainID,
		AccountNumber: solflareAccountNumber,
		Sequence:      signatures[0].Sequence,
		Address:       sdk.AccAddress(pubKeys[0].Address()).String(),
		PubKey:        &anypb.Any{TypeUrl: pubKeyAny.TypeUrl, Value: pubKeyAny.Value},
	}
	adaptable, ok := tx.(authsigning.V2AdaptableTx)
	require.True(t, ok)
	txData := adaptable.GetSigningTxData()

	signBytes := func() ([]byte, error) {
		return txConfig.SignModeHandler().GetSignBytes(
			ctx, signingv1beta1.SignMode_SIGN_MODE_SOLANA_TX_CARRIER, signerData, txData,
		)
	}

	require.NoError(t, testApp.SolanaCarrierKeeper.SetParams(ctx, solanacarrier.DefaultParams()))
	carried, err := signBytes()
	require.NoError(t, err)
	require.True(t, pubKeys[0].VerifySignature(carried, single.Signature),
		"the wallet's signature must verify over the carried message")

	withoutLighthouse := solanacarrier.DefaultParams()
	withoutLighthouse.AllowedPrograms = withoutLighthouse.AllowedPrograms[:1]
	require.Equal(t, "0306466fe5211732ffecadba72c39be7bc8ce5bbc5f7126b2c439b3a40000000",
		hex.EncodeToString(withoutLighthouse.AllowedPrograms[0].ProgramId), "only ComputeBudget, the v0.8.6 rule, is kept")
	require.NoError(t, testApp.SolanaCarrierKeeper.SetParams(ctx, withoutLighthouse))

	_, err = signBytes()
	require.ErrorContains(t, err, "instruction 3 invokes a program outside the allowlist")
}
