package solanacarrier

import (
	"bytes"
	"context"
	"slices"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"github.com/cosmos/btcutil/base58"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// SolanaCarrierTypeURL is the Any type URL of the carried Solana message.
	SolanaCarrierTypeURL = "/nolus.solanacarrier.v1.SolanaCarrier"

	ed25519PubKeyTypeURL = "/cosmos.crypto.ed25519.PubKey"

	// maxCarriedLen is the Solana packet cap the carried message must fit in.
	maxCarriedLen = 1232

	// minRequiredSignatures keeps account index 0 reserved for a fee payer the
	// signer is not. That slot must hold the governance-set fee payer, an
	// off-curve address no private key can produce a signature for, so a carrier
	// the wallet signs is always one signature short of a broadcastable Solana
	// transaction.
	minRequiredSignatures = 2
)

// memoProgramID is the 32-byte account key of the SPL Memo program
// (base58 MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr).
var memoProgramID = []byte{
	0x05, 0x4a, 0x53, 0x5a, 0x99, 0x29, 0x21, 0x06,
	0x4d, 0x24, 0xe8, 0x71, 0x60, 0xda, 0x38, 0x7c,
	0x7c, 0x35, 0xb5, 0xdd, 0xbc, 0x92, 0xbb, 0x81,
	0xe4, 0x1f, 0xa8, 0x40, 0x41, 0x05, 0x44, 0x8d,
}

var _ signing.SignModeHandler = SignModeHandler{}

// ParamsSource reports the module parameters a carrier is verified against: the
// fee payer at account index 0 and the program allowlist.
type ParamsSource interface {
	GetParams(ctx context.Context) (Params, error)
}

// SignModeHandler is the SIGN_MODE_SOLANA_TX_CARRIER implementation of
// signing.SignModeHandler.
type SignModeHandler struct {
	aminoJsonSignModeHandler *aminojson.SignModeHandler
	paramsSource             ParamsSource
}

type SignModeHandlerOptions struct {
	AminoJsonSignModeHandler *aminojson.SignModeHandler
	ParamsSource             ParamsSource
}

func NewSignModeHandler(options SignModeHandlerOptions) *SignModeHandler {
	return &SignModeHandler{
		aminoJsonSignModeHandler: options.AminoJsonSignModeHandler,
		paramsSource:             options.ParamsSource,
	}
}

func (s SignModeHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_SOLANA_TX_CARRIER
}

// GetSignBytes carries the Solana legacy message from the transaction's
// non-critical extension options, verifies it is bound to the amino JSON of the
// same transaction with the carrier stripped, and returns the carried message
// verbatim for the ed25519 signature to be verified over.
func (s SignModeHandler) GetSignBytes(ctx context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {
	carried, strippedTxData, err := extractCarrier(txData)
	if err != nil {
		return nil, err
	}

	// A foreign extension option surviving the strip makes this delegate call
	// fail-closed: amino JSON signing rejects any extension option.
	aminoJSON, err := s.aminoJsonSignModeHandler.GetSignBytes(ctx, signerData, strippedTxData)
	if err != nil {
		return nil, err
	}

	signerPubKey, err := signerEd25519PubKey(signerData)
	if err != nil {
		return nil, err
	}

	if len(carried) > maxCarriedLen {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER carried message is %d bytes, exceeds the %d-byte cap",
			len(carried), maxCarriedLen,
		)
	}

	message, err := ParseLegacyMessage(carried)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER cannot parse carried message: %v", err,
		)
	}

	if s.paramsSource == nil {
		return nil, sdkerrors.ErrLogic.Wrap(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER requires a params source",
		)
	}

	params, err := s.paramsSource.GetParams(ctx)
	if err != nil {
		return nil, paramsReadError{cause: err}
	}

	if err := validateBinding(message, signerPubKey, aminoJSON, params); err != nil {
		return nil, err
	}

	return carried, nil
}

// extractCarrier returns the carried message bytes and a copy of txData whose
// body has the single carrier extension removed and its bytes re-marshalled
// deterministically, leaving any foreign extension in place.
func extractCarrier(txData signing.TxData) ([]byte, signing.TxData, error) {
	body := txData.Body
	if body == nil {
		return nil, signing.TxData{}, sdkerrors.ErrInvalidRequest.Wrap(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER requires a transaction body",
		)
	}

	var carrierAny *anypb.Any
	kept := make([]*anypb.Any, 0, len(body.NonCriticalExtensionOptions))
	count := 0
	for _, ext := range body.NonCriticalExtensionOptions {
		if ext.GetTypeUrl() == SolanaCarrierTypeURL {
			carrierAny = ext
			count++
			continue
		}
		kept = append(kept, ext)
	}
	if count != 1 {
		return nil, signing.TxData{}, sdkerrors.ErrInvalidRequest.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER requires exactly one %s extension, found %d",
			SolanaCarrierTypeURL, count,
		)
	}

	var carrier SolanaCarrier
	if err := carrier.Unmarshal(carrierAny.GetValue()); err != nil {
		return nil, signing.TxData{}, sdkerrors.ErrInvalidRequest.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER cannot decode carrier extension: %v", err,
		)
	}

	strippedBody, ok := proto.Clone(body).(*txv1beta1.TxBody)
	if !ok {
		return nil, signing.TxData{}, sdkerrors.ErrInvalidRequest.Wrap(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER cannot clone transaction body",
		)
	}
	strippedBody.NonCriticalExtensionOptions = kept

	strippedBodyBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(strippedBody)
	if err != nil {
		return nil, signing.TxData{}, sdkerrors.ErrInvalidRequest.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER cannot re-marshal stripped body: %v", err,
		)
	}

	strippedTxData := signing.TxData{
		Body:                       strippedBody,
		AuthInfo:                   txData.AuthInfo,
		BodyBytes:                  strippedBodyBytes,
		AuthInfoBytes:              txData.AuthInfoBytes,
		BodyHasUnknownNonCriticals: txData.BodyHasUnknownNonCriticals,
	}
	return carrier.Message, strippedTxData, nil
}

// validateBinding enforces the carrier's fail-closed binding rules against the
// parsed message: account index 0 holds the governance-set fee payer, which is
// off curve and therefore unsignable, the signer is one of the other required
// signers, every instruction other than the SPL Memo invokes a program the
// allowlist names, within that program's instruction cap, and exactly one Memo
// instruction reproduces the transaction's amino JSON byte for byte. The pinned
// fee payer leaves a signed carrier with no signature for index 0, so a valid
// Nolus signature can never double as a broadcastable Solana transaction.
func validateBinding(message *LegacyMessage, signerPubKey, aminoJSON []byte, params Params) error {
	if int(message.NumRequiredSignatures) < minRequiredSignatures {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER requires at least %d required signatures, got %d",
			minRequiredSignatures, message.NumRequiredSignatures,
		)
	}
	if !bytes.Equal(params.FeePayer, message.AccountKeys[0]) {
		return sdkerrors.ErrInvalidRequest.Wrap(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER requires the configured fee payer at account index 0",
		)
	}
	if bytes.Equal(signerPubKey, message.AccountKeys[0]) {
		return sdkerrors.ErrInvalidRequest.Wrap(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER signer must not be the fee payer at account index 0",
		)
	}

	found := false
	for _, key := range message.AccountKeys[:message.NumRequiredSignatures] {
		if bytes.Equal(signerPubKey, key) {
			found = true
			break
		}
	}
	if !found {
		return sdkerrors.ErrInvalidRequest.Wrap(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER signer's key is not among the required signers",
		)
	}

	memos := 0
	// Indexed like params.AllowedPrograms, so caps are checked in params order.
	counts := make([]uint32, len(params.AllowedPrograms))
	for i := range message.Instructions {
		programID := message.AccountKeys[message.Instructions[i].ProgramIDIndex]
		if bytes.Equal(programID, memoProgramID) {
			memos++
			continue
		}
		j := slices.IndexFunc(params.AllowedPrograms, func(program AllowedProgram) bool {
			return bytes.Equal(program.ProgramId, programID)
		})
		if j < 0 {
			return sdkerrors.ErrInvalidRequest.Wrapf(
				"SignMode_SIGN_MODE_SOLANA_TX_CARRIER instruction %d invokes a program outside the allowlist",
				i,
			)
		}
		counts[j]++
	}
	for j, program := range params.AllowedPrograms {
		if counts[j] > program.MaxInstructions {
			return sdkerrors.ErrInvalidRequest.Wrapf(
				"SignMode_SIGN_MODE_SOLANA_TX_CARRIER allowlist tolerates at most %d instructions of program %s, found %d",
				program.MaxInstructions, base58.Encode(program.ProgramId), counts[j],
			)
		}
	}
	if memos != 1 {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER requires exactly one memo instruction bound to the transaction, found %d",
			memos,
		)
	}

	for i := range message.Instructions {
		if bytes.Equal(message.AccountKeys[message.Instructions[i].ProgramIDIndex], memoProgramID) &&
			!bytes.Equal(message.Instructions[i].Data, aminoJSON) {
			return sdkerrors.ErrInvalidRequest.Wrap(
				"SignMode_SIGN_MODE_SOLANA_TX_CARRIER memo instruction is not bound to the transaction's amino JSON",
			)
		}
	}

	return nil
}

func signerEd25519PubKey(signerData signing.SignerData) ([]byte, error) {
	if signerData.PubKey == nil {
		return nil, sdkerrors.ErrInvalidPubKey.Wrap(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER requires the signer's public key",
		)
	}
	if signerData.PubKey.TypeUrl != ed25519PubKeyTypeURL {
		return nil, sdkerrors.ErrInvalidPubKey.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_TX_CARRIER requires %q, got %q",
			ed25519PubKeyTypeURL, signerData.PubKey.TypeUrl,
		)
	}

	var pubKey ed25519.PubKey
	if err := pubKey.Unmarshal(signerData.PubKey.Value); err != nil {
		return nil, sdkerrors.ErrInvalidPubKey.Wrapf("cannot decode ed25519 public key: %v", err)
	}
	if len(pubKey.Key) != ed25519.PubKeySize {
		return nil, sdkerrors.ErrInvalidPubKey.Wrapf(
			"ed25519 public key must be %d bytes, got %d", ed25519.PubKeySize, len(pubKey.Key),
		)
	}

	return pubKey.Key, nil
}

var _ signing.SignModeHandler = (*SignModeHandler)(nil)

type paramsReadError struct {
	cause error
}

func (e paramsReadError) Error() string {
	return "SignMode_SIGN_MODE_SOLANA_TX_CARRIER cannot read the module params: " + e.cause.Error() + ": " + sdkerrors.ErrLogic.Error()
}

func (e paramsReadError) Cause() error {
	return sdkerrors.ErrLogic // the ABCI code and codespace are read through Cause
}

func (e paramsReadError) Unwrap() []error {
	return []error{sdkerrors.ErrLogic, e.cause} // errors.Is and errors.As reach both
}
