package solanacarrier

import (
	"bytes"

	"filippo.io/edwards25519"
	"github.com/cosmos/btcutil/base58"

	errorsmod "cosmossdk.io/errors"
)

const (
	maxAllowedPrograms         = 16
	maxProgramInstructions     = 64
	defaultProgramInstructions = 4
)

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

var computeBudgetProgramID = []byte{ // ComputeBudget111111111111111111111111111111
	0x03, 0x06, 0x46, 0x6f, 0xe5, 0x21, 0x17, 0x32,
	0xff, 0xec, 0xad, 0xba, 0x72, 0xc3, 0x9b, 0xe7,
	0xbc, 0x8c, 0xe5, 0xbb, 0xc5, 0xf7, 0x12, 0x6b,
	0x2c, 0x43, 0x9b, 0x3a, 0x40, 0x00, 0x00, 0x00,
}

var lighthouseProgramID = []byte{ // L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95
	0x04, 0xdf, 0xad, 0x79, 0x62, 0xff, 0xb1, 0xdd,
	0x92, 0x5d, 0x0a, 0x9f, 0xb5, 0xe6, 0xd0, 0x0c,
	0xe6, 0x19, 0x5b, 0xa8, 0xbb, 0x3a, 0x91, 0xfd,
	0x07, 0xef, 0x98, 0x60, 0xc5, 0xe9, 0x7b, 0xb8,
}

// DefaultParams returns the sentinel fee payer and an allowlist admitting up to
// four ComputeBudget and four Lighthouse instructions.
func DefaultParams() Params {
	return Params{
		FeePayer: bytes.Clone(DefaultFeePayer),
		AllowedPrograms: []AllowedProgram{
			{ProgramId: bytes.Clone(computeBudgetProgramID), MaxInstructions: defaultProgramInstructions},
			{ProgramId: bytes.Clone(lighthouseProgramID), MaxInstructions: defaultProgramInstructions},
		},
	}
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if err := validateFeePayer(p.FeePayer); err != nil {
		return err
	}

	return validateAllowedPrograms(p.AllowedPrograms)
}

func validateFeePayer(feePayer []byte) error {
	if len(feePayer) != pubKeyLen {
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

func validateAllowedPrograms(programs []AllowedProgram) error {
	if len(programs) > maxAllowedPrograms {
		return errorsmod.Wrapf(ErrInvalidAllowedPrograms,
			"at most %d programs may be allowed, got %d", maxAllowedPrograms, len(programs))
	}

	for i, program := range programs {
		if len(program.ProgramId) != pubKeyLen {
			return errorsmod.Wrapf(ErrInvalidAllowedPrograms,
				"entry %d: program id must be %d bytes, got %d", i, pubKeyLen, len(program.ProgramId))
		}
		// The memo is the binding itself: exactly one, carrying the amino JSON.
		if bytes.Equal(program.ProgramId, memoProgramID) {
			return errorsmod.Wrapf(ErrInvalidAllowedPrograms,
				"entry %d: the SPL Memo program is always bound and cannot be listed", i)
		}
		if program.MaxInstructions < 1 || program.MaxInstructions > maxProgramInstructions {
			return errorsmod.Wrapf(ErrInvalidAllowedPrograms,
				"entry %d: program %s max instructions must be between 1 and %d, got %d",
				i, base58.Encode(program.ProgramId), maxProgramInstructions, program.MaxInstructions)
		}
		for _, earlier := range programs[:i] {
			if bytes.Equal(earlier.ProgramId, program.ProgramId) {
				return errorsmod.Wrapf(ErrInvalidAllowedPrograms,
					"entry %d: program %s is listed more than once", i, base58.Encode(program.ProgramId))
			}
		}
	}

	return nil
}
