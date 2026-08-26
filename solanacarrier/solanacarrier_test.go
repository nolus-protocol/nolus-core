package solanacarrier_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
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
	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// Carrier captured 2026-08-25 from Phantom + Ledger. Phantom injected two
// ComputeBudget instructions before signing, so the chain sees four instructions
// where the client built two; see .claude/spike/fixtures-solflare-ocms.md.
const (
	phantomPubKeyHex = "54a18bb661d66d5a01e0fa06f3588ec0b12c81c3cc39263748fa2e9ff99acf4a"

	phantomCarrierMessageHex = "0200030548ab05fd4f9c5a20ee631f7f45e8ac8ecf133b2ab59b247b452b6ea5dd5bf94854a18bb661d66d5a01e0fa06f3588ec0b12c81c3cc39263748fa2e9ff99acf4a00000000000000000000000000000000000000000000000000000000000000000306466fe5211732ffecadba72c39be7bc8ce5bbc5f7126b2c439b3a40000000054a535a992921064d24e87160da387c7c35b5ddbc92bb81e41fa8404105448dd3266180b58ab52ad71ace42190f053f25dd7e4b74dba3a6c07aa8f7ec52201c04030009031cd604000000000003000502529c0300020201010c020000000000000000000000040101ef037b226163636f756e745f6e756d626572223a22313834343637222c22636861696e5f6964223a226e6f6c75732d31222c22666565223a7b22616d6f756e74223a5b7b22616d6f756e74223a2233303030222c2264656e6f6d223a22756e6c73227d5d2c22676173223a22353030303030227d2c226d656d6f223a22222c226d736773223a5b7b2274797065223a227761736d2f4d736745786563757465436f6e7472616374222c2276616c7565223a7b22636f6e7472616374223a226e6f6c757331776e36323573346a636d766b30737a706c3835726a35617a6b6663367375797666373571367672646473636a6470687476653873356767343266222c2266756e6473223a5b7b22616d6f756e74223a22313530303030303030222c2264656e6f6d223a226962632f46344237433146314537423143423031384545304642413539394130413438413030383131433044314442314542434538324430423034423445343045344545227d5d2c226d7367223a7b226f70656e5f6c65617365223a7b2263757272656e6379223a22534f4c222c226d61785f6c7464223a3630307d7d2c2273656e646572223a226e6f6c7573313064303779323635676d6d757674347a30773961773838306a6e73723730306a766a7236356b227d7d5d2c2273657175656e6365223a223432227d"

	phantomSignatureHex = "29f29a96b74c440b4218c15f81be386170686fc84bd687a1d5e2e6fc34420fc575320402eccdda167e18c60a1a385b6d20506fbcc4413d189225184261d35d0f"
)

// The Nolus transaction the carrier's memo is bound to.
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

// maxCarriedLen is the Solana packet cap the carried message must fit in.
const maxCarriedLen = 1232

const signModeSolanaTxCarrier = signingv1beta1.SignMode(193)

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

func newHandler(t *testing.T) *solanacarrier.SignModeHandler {
	t.Helper()
	return solanacarrier.NewSignModeHandler(solanacarrier.SignModeHandlerOptions{
		AminoJsonSignModeHandler: newAminoHandler(t),
	})
}

func leaseOpenMsg() sdk.Msg {
	return &wasmtypes.MsgExecuteContract{
		Sender:   fixtureSender,
		Contract: fixtureContract,
		Msg:      []byte(fixtureExecuteMsg),
		Funds:    sdk.NewCoins(sdk.NewCoin(fixtureFundsDenom, math.NewInt(fixtureFundsAmount))),
	}
}

func txDataWithExtensions(t *testing.T, extensions ...*codectypes.Any) txsigning.TxData {
	t.Helper()
	builder, ok := app.MakeEncodingConfig(app.ModuleBasics).TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	require.True(t, ok, "tx builder must accept non-critical extension options")

	require.NoError(t, builder.SetMsgs(leaseOpenMsg()))
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin(fixtureFeeDenom, fixtureFeeAmount)))
	builder.SetGasLimit(fixtureGasLimit)
	builder.SetMemo("")
	builder.SetNonCriticalExtensionOptions(extensions...)

	adaptable, ok := builder.GetTx().(authsigning.V2AdaptableTx)
	require.True(t, ok, "tx builder must produce a tx carrying x/tx signing data")
	return adaptable.GetSigningTxData()
}

func carrierExtension(t *testing.T, message []byte) *codectypes.Any {
	t.Helper()
	extension, err := codectypes.NewAnyWithValue(&solanacarrier.SolanaCarrier{Message: message})
	require.NoError(t, err)
	return extension
}

func foreignExtension(t *testing.T) *codectypes.Any {
	t.Helper()
	extension, err := codectypes.NewAnyWithValue(&banktypes.MsgSend{
		FromAddress: fixtureSender,
		ToAddress:   fixtureSender,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(fixtureFeeDenom, 1)),
	})
	require.NoError(t, err)
	return extension
}

func signerDataFor(t *testing.T, pubKey cryptotypes.PubKey) txsigning.SignerData {
	t.Helper()
	anyPubKey, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	return txsigning.SignerData{
		ChainID:       fixtureChainID,
		AccountNumber: fixtureAccountNumber,
		Sequence:      fixtureSequence,
		Address:       fixtureSender,
		PubKey:        &anypb.Any{TypeUrl: anyPubKey.TypeUrl, Value: anyPubKey.Value},
	}
}

// boundAminoJSON is the payload the carrier's memo instruction must reproduce
// byte for byte: the amino JSON of the transaction with the carrier stripped.
func boundAminoJSON(t *testing.T) []byte {
	t.Helper()
	content, err := newAminoHandler(t).GetSignBytes(
		context.Background(),
		signerDataFor(t, ed25519PubKeyFromHex(t, phantomPubKeyHex)),
		txDataWithExtensions(t),
	)
	require.NoError(t, err)
	return content
}

func signBytesFor(t *testing.T, pubKey cryptotypes.PubKey, extensions ...*codectypes.Any) ([]byte, error) {
	t.Helper()
	return newHandler(t).GetSignBytes(
		context.Background(),
		signerDataFor(t, pubKey),
		txDataWithExtensions(t, extensions...),
	)
}

func TestModeIsSolanaTxCarrier(t *testing.T) {
	require.Equal(t, signModeSolanaTxCarrier, newHandler(t).Mode())
}

func TestGetSignBytesReturnsPhantomCarriedMessageVerbatim(t *testing.T) {
	carried := mustDecodeHex(t, phantomCarrierMessageHex)

	got, err := signBytesFor(t, ed25519PubKeyFromHex(t, phantomPubKeyHex), carrierExtension(t, carried))
	require.NoError(t, err)
	require.Equal(t, carried, got)
}

func TestPhantomFixtureSignatureVerifiesOverSignBytes(t *testing.T) {
	pubKey := ed25519PubKeyFromHex(t, phantomPubKeyHex)

	got, err := signBytesFor(t, pubKey, carrierExtension(t, mustDecodeHex(t, phantomCarrierMessageHex)))
	require.NoError(t, err)
	require.True(t,
		ed25519.Verify(ed25519.PublicKey(pubKey.Bytes()), got, mustDecodeHex(t, phantomSignatureHex)),
		"the signature Phantom produced must verify over the sign bytes the chain returns",
	)
}

func TestGetSignBytesAcceptsHandBuiltCarrierMessage(t *testing.T) {
	signerKey := mustDecodeHex(t, phantomPubKeyHex)
	carried := carrierMessage(signerKey, boundAminoJSON(t)).encode()

	got, err := signBytesFor(t, ed25519PubKeyFromHex(t, phantomPubKeyHex), carrierExtension(t, carried))
	require.NoError(t, err)
	require.Equal(t, carried, got)
}

func TestGetSignBytesToleratesUnrecognisedInstructions(t *testing.T) {
	signerKey := mustDecodeHex(t, phantomPubKeyHex)
	message := carrierMessage(signerKey, boundAminoJSON(t))
	message.instructions = append(message.instructions, solanaInstruction{
		programIDIndex: systemProgramIndex,
		data:           []byte("an instruction the chain knows nothing about"),
	})
	carried := message.encode()

	got, err := signBytesFor(t, ed25519PubKeyFromHex(t, phantomPubKeyHex), carrierExtension(t, carried))
	require.NoError(t, err)
	require.Equal(t, carried, got)
}

func TestGetSignBytesAcceptsMessageAtPacketCap(t *testing.T) {
	signerKey := mustDecodeHex(t, phantomPubKeyHex)
	carried := padToLength(t, carrierMessage(signerKey, boundAminoJSON(t)), maxCarriedLen).encode()
	require.Len(t, carried, maxCarriedLen)

	got, err := signBytesFor(t, ed25519PubKeyFromHex(t, phantomPubKeyHex), carrierExtension(t, carried))
	require.NoError(t, err)
	require.Equal(t, carried, got)
}

func TestGetSignBytesRejectsInvalidCarriedMessage(t *testing.T) {
	signerKey := mustDecodeHex(t, phantomPubKeyHex)

	tests := []struct {
		name    string
		carried func(t *testing.T) []byte
	}{
		{
			name: "memo json differs by one byte",
			carried: func(t *testing.T) []byte {
				t.Helper()
				tampered := bytes.Replace(boundAminoJSON(t), []byte(`"sequence":"42"`), []byte(`"sequence":"43"`), 1)
				require.NotEqual(t, boundAminoJSON(t), tampered, "the tamper must actually change the memo")
				return carrierMessage(signerKey, tampered).encode()
			},
		},
		{
			name: "no memo instruction",
			carried: func(t *testing.T) []byte {
				t.Helper()
				message := carrierMessage(signerKey, boundAminoJSON(t))
				message.instructions = message.instructions[:len(message.instructions)-1]
				return message.encode()
			},
		},
		{
			name: "two matching memo instructions",
			carried: func(t *testing.T) []byte {
				t.Helper()
				message := carrierMessage(signerKey, boundAminoJSON(t))
				message.instructions = append(message.instructions, message.instructions[len(message.instructions)-1])
				return message.encode()
			},
		},
		{
			name: "signer not among required signers",
			carried: func(t *testing.T) []byte {
				t.Helper()
				message := carrierMessage(signerKey, boundAminoJSON(t))
				message.numRequiredSignatures = 1
				return message.encode()
			},
		},
		{
			name: "signer key absent from the message",
			carried: func(t *testing.T) []byte {
				t.Helper()
				message := carrierMessage(bytes.Repeat([]byte{0x11}, 32), boundAminoJSON(t))
				return message.encode()
			},
		},
		{
			name: "message one byte past the packet cap",
			carried: func(t *testing.T) []byte {
				t.Helper()
				return padToLength(t, carrierMessage(signerKey, boundAminoJSON(t)), maxCarriedLen+1).encode()
			},
		},
		{
			name: "versioned message header",
			carried: func(t *testing.T) []byte {
				t.Helper()
				return append([]byte{0x80}, carrierMessage(signerKey, boundAminoJSON(t)).encode()...)
			},
		},
		{
			name: "truncated message",
			carried: func(t *testing.T) []byte {
				t.Helper()
				encoded := carrierMessage(signerKey, boundAminoJSON(t)).encode()
				return encoded[:len(encoded)-1]
			},
		},
		{
			name: "trailing bytes after the last instruction",
			carried: func(t *testing.T) []byte {
				t.Helper()
				return append(carrierMessage(signerKey, boundAminoJSON(t)).encode(), 0x00)
			},
		},
		{
			name: "garbage bytes",
			carried: func(t *testing.T) []byte {
				t.Helper()
				return []byte{0xde, 0xad, 0xbe, 0xef}
			},
		},
		{
			name: "empty message",
			carried: func(t *testing.T) []byte {
				t.Helper()
				return nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := signBytesFor(t, ed25519PubKeyFromHex(t, phantomPubKeyHex), carrierExtension(t, tc.carried(t)))
			require.Error(t, err)
		})
	}
}

func TestGetSignBytesRejectsInvalidCarrierExtensions(t *testing.T) {
	signerKey := mustDecodeHex(t, phantomPubKeyHex)

	tests := []struct {
		name       string
		extensions func(t *testing.T) []*codectypes.Any
	}{
		{
			name: "no carrier extension",
			extensions: func(t *testing.T) []*codectypes.Any {
				t.Helper()
				return nil
			},
		},
		{
			name: "two carrier extensions",
			extensions: func(t *testing.T) []*codectypes.Any {
				t.Helper()
				carried := carrierMessage(signerKey, boundAminoJSON(t)).encode()
				return []*codectypes.Any{carrierExtension(t, carried), carrierExtension(t, carried)}
			},
		},
		{
			name: "foreign extension alongside the carrier",
			extensions: func(t *testing.T) []*codectypes.Any {
				t.Helper()
				carried := carrierMessage(signerKey, boundAminoJSON(t)).encode()
				return []*codectypes.Any{carrierExtension(t, carried), foreignExtension(t)}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := signBytesFor(t, ed25519PubKeyFromHex(t, phantomPubKeyHex), tc.extensions(t)...)
			require.Error(t, err)
		})
	}
}

func TestGetSignBytesRejectsNonEd25519PubKey(t *testing.T) {
	signerKey := mustDecodeHex(t, phantomPubKeyHex)
	carried := carrierMessage(signerKey, boundAminoJSON(t)).encode()

	_, err := signBytesFor(t, secp256k1.GenPrivKey().PubKey(), carrierExtension(t, carried))
	require.Error(t, err, "required signers hold 32-byte ed25519 keys; other key types cannot match one")
}

func TestGetSignBytesRejectsSignerWhoseKeyIsNotInTheMessage(t *testing.T) {
	signerKey := mustDecodeHex(t, phantomPubKeyHex)
	carried := carrierMessage(signerKey, boundAminoJSON(t)).encode()

	_, err := signBytesFor(t, sdked25519.GenPrivKey().PubKey(), carrierExtension(t, carried))
	require.Error(t, err, "a signer absent from the carried message's required signers must be rejected")
}
