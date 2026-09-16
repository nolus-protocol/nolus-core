package solanacarrier

import (
	"bytes"

	"filippo.io/edwards25519"

	errorsmod "cosmossdk.io/errors"
)

// feePayerLen is the length of a Solana account key.
const feePayerLen = 32

// DefaultFeePayer is the sentinel account key placed at account index 0 of every
// carrier, base58 EvysLTsKMWc2A9kNUBaeJF99kvZvPLjfbow2Yk4gQYXZ. It is a
// program-derived address: off the ed25519 curve, so no private key exists for
// it and the fee-payer signature a broadcastable Solana transaction needs can
// never be produced.
var DefaultFeePayer = []byte{
	0xce, 0xfc, 0x03, 0x7c, 0x11, 0x43, 0x6d, 0xc8,
	0x5b, 0x69, 0x43, 0xb4, 0x3c, 0x94, 0xec, 0x50,
	0xc0, 0x49, 0x36, 0x96, 0xcb, 0x7e, 0x79, 0x7e,
	0x84, 0xc1, 0x76, 0x30, 0x38, 0xe9, 0xad, 0x70,
}

// DefaultParams returns the default solanacarrier module parameters.
func DefaultParams() Params {
	return Params{
		FeePayer: bytes.Clone(DefaultFeePayer),
	}
}

// Validate validates the set of params.
func (p Params) Validate() error {
	return validateFeePayer(p.FeePayer)
}

func validateFeePayer(feePayer []byte) error {
	if len(feePayer) != feePayerLen {
		return errorsmod.Wrapf(ErrInvalidFeePayer, "fee payer must be 32 bytes, got %d", len(feePayer))
	}

	// A decodable point is a spendable account: someone can hold its private key
	// and sign for the fee-payer slot, turning a carrier into a broadcastable
	// Solana transaction. Only program-derived, off-curve keys are accepted.
	if _, err := new(edwards25519.Point).SetBytes(feePayer); err == nil {
		return errorsmod.Wrap(ErrInvalidFeePayer, "fee payer is on-curve; it must be an off-curve program-derived address")
	}

	return nil
}
