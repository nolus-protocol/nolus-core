package v2

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	types "github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// ParamsKeeper reads and writes the x/solanacarrier params; SetParams validates.
type ParamsKeeper interface {
	GetParams(ctx context.Context) (types.Params, error)
	SetParams(ctx context.Context, params types.Params) error
}

// AllowedPrograms returns the program allowlist the v1 to v2 migration writes:
// ComputeBudget and Lighthouse, at most four instructions each.
func AllowedPrograms() []types.AllowedProgram {
	return []types.AllowedProgram{
		{
			// ComputeBudget111111111111111111111111111111
			ProgramId: []byte{
				0x03, 0x06, 0x46, 0x6f, 0xe5, 0x21, 0x17, 0x32,
				0xff, 0xec, 0xad, 0xba, 0x72, 0xc3, 0x9b, 0xe7,
				0xbc, 0x8c, 0xe5, 0xbb, 0xc5, 0xf7, 0x12, 0x6b,
				0x2c, 0x43, 0x9b, 0x3a, 0x40, 0x00, 0x00, 0x00,
			},
			MaxInstructions: 4,
		},
		{
			// L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95
			ProgramId: []byte{
				0x04, 0xdf, 0xad, 0x79, 0x62, 0xff, 0xb1, 0xdd,
				0x92, 0x5d, 0x0a, 0x9f, 0xb5, 0xe6, 0xd0, 0x0c,
				0xe6, 0x19, 0x5b, 0xa8, 0xbb, 0x3a, 0x91, 0xfd,
				0x07, 0xef, 0x98, 0x60, 0xc5, 0xe9, 0x7b, 0xb8,
			},
			MaxInstructions: 4,
		},
	}
}

// Migrate migrates the x/solanacarrier module state from consensus version 1 to
// version 2: the stored params keep their fee payer and gain AllowedPrograms.
// Params missing from the store, or failing validation, are an error; nothing is
// written then.
func Migrate(ctx context.Context, k ParamsKeeper) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return errorsmod.Wrap(err, "read solanacarrier v1 params")
	}

	params.AllowedPrograms = AllowedPrograms()

	return errorsmod.Wrap(k.SetParams(ctx, params), "write solanacarrier v2 params")
}
