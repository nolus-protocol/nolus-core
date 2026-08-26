package solanaoffchain

import (
	"context"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const ed25519PubKeyTypeURL = "/cosmos.crypto.ed25519.PubKey"

const (
	offchainMessageSigningDomain = "solana offchain"

	offchainMessageVersion     byte = 0x01
	offchainMessageSignerCount byte = 0x01

	envelopePreambleLen = 1 + len(offchainMessageSigningDomain) + 1 + 1 + ed25519.PubKeySize

	maxContentLen = 1232 - envelopePreambleLen
)

var _ signing.SignModeHandler = SignModeHandler{}

// SignModeHandler is the SIGN_MODE_SOLANA_OFFCHAIN implementation of signing.SignModeHandler.
type SignModeHandler struct {
	aminoJsonSignModeHandler *aminojson.SignModeHandler
}

type SignModeHandlerOptions struct {
	AminoJsonSignModeHandler *aminojson.SignModeHandler
}

func (s SignModeHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_SOLANA_OFFCHAIN
}

func NewSignModeHandler(options SignModeHandlerOptions) *SignModeHandler {
	h := &SignModeHandler{
		aminoJsonSignModeHandler: options.AminoJsonSignModeHandler,
	}
	return h
}

func (s SignModeHandler) GetSignBytes(ctx context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {
	content, err := s.aminoJsonSignModeHandler.GetSignBytes(ctx, signerData, txData)
	if err != nil {
		return nil, err
	}

	pubKey, err := signerEd25519PubKey(signerData)
	if err != nil {
		return nil, err
	}

	if len(content) > maxContentLen {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_OFFCHAIN content is %d bytes, exceeds the %d-byte cap",
			len(content), maxContentLen,
		)
	}

	envelope := make([]byte, 0, envelopePreambleLen+len(content))
	envelope = append(envelope, 0xff)
	envelope = append(envelope, offchainMessageSigningDomain...)
	envelope = append(envelope, offchainMessageVersion, offchainMessageSignerCount)
	envelope = append(envelope, pubKey...)
	envelope = append(envelope, content...)

	return envelope, nil
}

func signerEd25519PubKey(signerData signing.SignerData) ([]byte, error) {
	if signerData.PubKey == nil {
		return nil, sdkerrors.ErrInvalidPubKey.Wrap("SignMode_SIGN_MODE_SOLANA_OFFCHAIN requires the signer's public key")
	}
	if signerData.PubKey.TypeUrl != ed25519PubKeyTypeURL {
		return nil, sdkerrors.ErrInvalidPubKey.Wrapf(
			"SignMode_SIGN_MODE_SOLANA_OFFCHAIN requires %q, got %q",
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
