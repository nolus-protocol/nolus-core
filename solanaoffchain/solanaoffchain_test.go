package solanaoffchain_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"strings"
	"testing"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/math"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdked25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/solanaoffchain"
)

// Fixtures captured from the Solflare extension's wallet-standard
// solana:signOffchainMessage feature on 2026-08-25 and independently verified;
// see .claude/spike/fixtures-solflare-ocms.md.
const (
	solflarePubKeyHex = "8dbdef2ce14b74d17b2daae9881a7ebe480a1a770161e6423d22a4f0450667f2"

	solflareEnvelopeHex = "ff736f6c616e61206f6666636861696e01018dbdef2ce14b74d17b2daae9881a7ebe480a1a770161e6423d22a4f0450667f27b226163636f756e745f6e756d626572223a22313834343637222c22636861696e5f6964223a226e6f6c75732d31222c22666565223a7b22616d6f756e74223a5b7b22616d6f756e74223a2233303030222c2264656e6f6d223a22756e6c73227d5d2c22676173223a22353030303030227d2c226d656d6f223a22222c226d736773223a5b7b2274797065223a227761736d2f4d736745786563757465436f6e7472616374222c2276616c7565223a7b22636f6e7472616374223a226e6f6c757331776e36323573346a636d766b30737a706c3835726a35617a6b6663367375797666373571367672646473636a6470687476653873356767343266222c2266756e6473223a5b7b22616d6f756e74223a22313530303030303030222c2264656e6f6d223a226962632f46344237433146314537423143423031384545304642413539394130413438413030383131433044314442314542434538324430423034423445343045344545227d5d2c226d7367223a7b226f70656e5f6c65617365223a7b2263757272656e6379223a22534f4c222c226d61785f6c7464223a3630307d7d2c2273656e646572223a226e6f6c7573313064303779323635676d6d757674347a30773961773838306a6e73723730306a766a7236356b227d7d5d2c2273657175656e6365223a223432227d"

	solflareSignatureHex = "151ef15903375e972d96c9be22658a771dfe8059183efb454262a63de1369c8c9977fd1332535c52d5b7ab9f6631458e8e1d23a73d03709c6c4fa047ceeac806"
)

// The Nolus transaction the captured envelope was signed over.
const (
	fixtureChainID       = "nolus-1"
	fixtureAccountNumber = 184467
	fixtureSequence      = 42
	fixtureSender        = "nolus10d07y265gmmuvt4z0w9aw880jnsr700jvjr65k"
	fixtureContract      = "nolus1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s5gg42f"
	fixtureFundsDenom    = "ibc/F4B7C1F1E7B1CB018EE0FBA599A0A48A00811C0D1DB1EBCE82D0B04B4E40E4EE"
	fixtureFundsAmount   = 150000000
	fixtureExecuteMsg    = `{"open_lease":{"currency":"SOL","max_ltd":600}}`
	fixtureFeeDenom      = "unls"
	fixtureFeeAmount     = 3000
	fixtureGasLimit      = 500000
)

// sRFC-38 v1: 0xff || "solana offchain" || version || signer count || 32-byte pubkey.
const (
	envelopePreambleLen = 16 + 1 + 1 + 32

	// The Ledger Solana app caps the whole signed message at 1232 bytes.
	maxContentLen = 1232 - envelopePreambleLen
)

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	bz, err := hex.DecodeString(s)
	require.NoError(t, err)
	return bz
}

func ed25519PubKeyFromHex(t *testing.T, s string) cryptotypes.PubKey {
	t.Helper()
	return &sdked25519.PubKey{Key: mustDecodeHex(t, s)}
}

func newAminoHandler(t *testing.T) *aminojson.SignModeHandler {
	t.Helper()
	encodingConfig := app.MakeEncodingConfig(app.ModuleBasics)
	signingOpts, err := authtx.NewDefaultSigningOptions()
	require.NoError(t, err)
	signingOpts.FileResolver = encodingConfig.InterfaceRegistry

	return aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
		FileResolver: signingOpts.FileResolver,
		TypeResolver: signingOpts.TypeResolver,
	})
}

func newHandler(t *testing.T) *solanaoffchain.SignModeHandler {
	t.Helper()
	return solanaoffchain.NewSignModeHandler(solanaoffchain.SignModeHandlerOptions{
		AminoJsonSignModeHandler: newAminoHandler(t),
	})
}

func txDataFor(t *testing.T, memo string, msgs ...sdk.Msg) txsigning.TxData {
	t.Helper()
	builder := app.MakeEncodingConfig(app.ModuleBasics).TxConfig.NewTxBuilder()
	require.NoError(t, builder.SetMsgs(msgs...))
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin(fixtureFeeDenom, fixtureFeeAmount)))
	builder.SetGasLimit(fixtureGasLimit)
	builder.SetMemo(memo)

	adaptable, ok := builder.GetTx().(authsigning.V2AdaptableTx)
	require.True(t, ok, "tx builder must produce a tx carrying x/tx signing data")
	return adaptable.GetSigningTxData()
}

func leaseOpenTxData(t *testing.T, memo string) txsigning.TxData {
	t.Helper()
	return txDataFor(t, memo, &wasmtypes.MsgExecuteContract{
		Sender:   fixtureSender,
		Contract: fixtureContract,
		Msg:      []byte(fixtureExecuteMsg),
		Funds:    sdk.NewCoins(sdk.NewCoin(fixtureFundsDenom, math.NewInt(fixtureFundsAmount))),
	})
}

func bankSendTxData(t *testing.T, memo string) txsigning.TxData {
	t.Helper()
	return txDataFor(t, memo, &banktypes.MsgSend{
		FromAddress: fixtureSender,
		ToAddress:   fixtureSender,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(fixtureFeeDenom, 1)),
	})
}

func signerDataFor(t *testing.T, pubKey cryptotypes.PubKey) txsigning.SignerData {
	t.Helper()
	signerData := txsigning.SignerData{
		ChainID:       fixtureChainID,
		AccountNumber: fixtureAccountNumber,
		Sequence:      fixtureSequence,
		Address:       fixtureSender,
	}
	if pubKey == nil {
		return signerData
	}

	anyPubKey, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)
	signerData.PubKey = &anypb.Any{TypeUrl: anyPubKey.TypeUrl, Value: anyPubKey.Value}
	return signerData
}

func aminoContentLen(t *testing.T, memo string) int {
	t.Helper()
	content, err := newAminoHandler(t).GetSignBytes(
		context.Background(),
		signerDataFor(t, ed25519PubKeyFromHex(t, solflarePubKeyHex)),
		leaseOpenTxData(t, memo),
	)
	require.NoError(t, err)
	return len(content)
}

func TestModeIsSolanaOffchain(t *testing.T) {
	require.Equal(t, signingv1beta1.SignMode_SIGN_MODE_SOLANA_OFFCHAIN, newHandler(t).Mode())
}

func TestGetSignBytesMatchesSolflareEnvelopeFixture(t *testing.T) {
	pubKey := ed25519PubKeyFromHex(t, solflarePubKeyHex)

	got, err := newHandler(t).GetSignBytes(context.Background(), signerDataFor(t, pubKey), leaseOpenTxData(t, ""))
	require.NoError(t, err)
	require.Equal(t, mustDecodeHex(t, solflareEnvelopeHex), got)
}

func TestSolflareFixtureSignatureVerifiesOverSignBytes(t *testing.T) {
	pubKey := ed25519PubKeyFromHex(t, solflarePubKeyHex)

	got, err := newHandler(t).GetSignBytes(context.Background(), signerDataFor(t, pubKey), leaseOpenTxData(t, ""))
	require.NoError(t, err)
	require.True(t,
		ed25519.Verify(ed25519.PublicKey(pubKey.Bytes()), got, mustDecodeHex(t, solflareSignatureHex)),
		"the signature Solflare produced must verify over the sign bytes the chain reconstructs",
	)
}

func TestGetSignBytesFramesAminoJSONInV1Envelope(t *testing.T) {
	pubKey := sdked25519.GenPrivKey().PubKey()
	signerData := signerDataFor(t, pubKey)
	txData := bankSendTxData(t, "")

	content, err := newAminoHandler(t).GetSignBytes(context.Background(), signerData, txData)
	require.NoError(t, err)

	want := append([]byte{0xff}, []byte("solana offchain")...)
	want = append(want, 0x01, 0x01)
	want = append(want, pubKey.Bytes()...)
	want = append(want, content...)

	got, err := newHandler(t).GetSignBytes(context.Background(), signerData, txData)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetSignBytesRejectsNonEd25519PubKey(t *testing.T) {
	signerData := signerDataFor(t, secp256k1.GenPrivKey().PubKey())

	_, err := newHandler(t).GetSignBytes(context.Background(), signerData, leaseOpenTxData(t, ""))
	require.Error(t, err, "the envelope binds a 32-byte ed25519 key; other key types have no place in it")
}

func TestGetSignBytesRejectsMissingPubKey(t *testing.T) {
	signerData := signerDataFor(t, nil)

	_, err := newHandler(t).GetSignBytes(context.Background(), signerData, leaseOpenTxData(t, ""))
	require.Error(t, err, "the envelope cannot be built without the signer's public key")
}

func TestGetSignBytesAcceptsMaxLengthContent(t *testing.T) {
	pad := maxContentLen - aminoContentLen(t, "")
	require.Positive(t, pad, "the fixture tx must be short enough to pad up to the cap")

	got, err := newHandler(t).GetSignBytes(
		context.Background(),
		signerDataFor(t, ed25519PubKeyFromHex(t, solflarePubKeyHex)),
		leaseOpenTxData(t, strings.Repeat("a", pad)),
	)
	require.NoError(t, err)
	require.Len(t, got, envelopePreambleLen+maxContentLen)
}

func TestGetSignBytesRejectsContentOverMaxLength(t *testing.T) {
	pad := maxContentLen - aminoContentLen(t, "") + 1

	_, err := newHandler(t).GetSignBytes(
		context.Background(),
		signerDataFor(t, ed25519PubKeyFromHex(t, solflarePubKeyHex)),
		leaseOpenTxData(t, strings.Repeat("a", pad)),
	)
	require.Error(t, err, "content one byte past the cap must be rejected")
}

func TestGetSignBytesRejectsTxWithExtensionOptions(t *testing.T) {
	builder, ok := app.MakeEncodingConfig(app.ModuleBasics).TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	require.True(t, ok, "tx builder must accept extension options")
	require.NoError(t, builder.SetMsgs(&banktypes.MsgSend{
		FromAddress: fixtureSender,
		ToAddress:   fixtureSender,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(fixtureFeeDenom, 1)),
	}))
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin(fixtureFeeDenom, fixtureFeeAmount)))
	builder.SetGasLimit(fixtureGasLimit)

	foreign, err := codectypes.NewAnyWithValue(&banktypes.MsgSend{
		FromAddress: fixtureSender,
		ToAddress:   fixtureSender,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(fixtureFeeDenom, 1)),
	})
	require.NoError(t, err)
	builder.SetNonCriticalExtensionOptions(foreign)

	adaptable, ok := builder.GetTx().(authsigning.V2AdaptableTx)
	require.True(t, ok, "tx builder must produce a tx carrying x/tx signing data")

	_, err = newHandler(t).GetSignBytes(
		context.Background(),
		signerDataFor(t, ed25519PubKeyFromHex(t, solflarePubKeyHex)),
		adaptable.GetSigningTxData(),
	)
	require.Error(t, err, "mode 192 wraps the amino JSON handler, which does not support extension options")
}
