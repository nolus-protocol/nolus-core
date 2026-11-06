package solanacarrier_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func TestParamsValidateDefault(t *testing.T) {
	require.NoError(t, solanacarrier.DefaultParams().Validate())
	require.Len(t, solanacarrier.DefaultFeePayer, 32)
}

func TestParamsValidateWrongLength(t *testing.T) {
	for name, feePayer := range map[string][]byte{
		"nil":       nil,
		"empty":     {},
		"31 bytes":  solanacarrier.DefaultFeePayer[:31],
		"33 bytes":  append(append([]byte{}, solanacarrier.DefaultFeePayer...), 0x00),
		"truncated": solanacarrier.DefaultFeePayer[:16],
	} {
		t.Run(name, func(t *testing.T) {
			err := solanacarrier.Params{FeePayer: feePayer}.Validate()
			require.ErrorIs(t, err, solanacarrier.ErrInvalidFeePayer)
			require.Contains(t, err.Error(), "32 bytes")
		})
	}
}

func TestParamsValidateRejectsOnCurveFeePayer(t *testing.T) {
	// The SPL Memo program id is a plain on-curve ed25519 point.
	onCurve := mustHex(t, "054a535a992921064d24e87160da387c7c35b5ddbc92bb81e41fa8404105448d")

	err := solanacarrier.Params{FeePayer: onCurve}.Validate()
	require.ErrorIs(t, err, solanacarrier.ErrInvalidFeePayer)
	require.True(t, strings.Contains(err.Error(), "on-curve"), "error must name on-curve, got %q", err)
}

func TestDefaultFeePayerIsTheSentinelPDA(t *testing.T) {
	require.Equal(t,
		mustHex(t, "cefc037c11436dc85b6943b43c94ec50c0493696cb7e797e84c1763038e9ad70"),
		solanacarrier.DefaultFeePayer,
	)
}

func TestDefaultParamsAllowComputeBudgetAndLighthouse(t *testing.T) {
	require.Equal(t, []solanacarrier.AllowedProgram{
		{ProgramId: mustHex(t, "0306466fe5211732ffecadba72c39be7bc8ce5bbc5f7126b2c439b3a40000000"), MaxInstructions: 4},
		{ProgramId: mustHex(t, "04dfad7962ffb1dd925d0a9fb5e6d00ce6195ba8bb3a91fd07ef9860c5e97bb8"), MaxInstructions: 4},
	}, solanacarrier.DefaultParams().AllowedPrograms)
}

func distinctPrograms(n int, maxInstructions uint32) []solanacarrier.AllowedProgram {
	programs := make([]solanacarrier.AllowedProgram, n)
	for i := range programs {
		programs[i] = solanacarrier.AllowedProgram{
			ProgramId:       bytes.Repeat([]byte{byte(0x20 + i)}, 32),
			MaxInstructions: maxInstructions,
		}
	}
	return programs
}

func TestParamsValidateAllowedPrograms(t *testing.T) {
	memoProgram := mustHex(t, "054a535a992921064d24e87160da387c7c35b5ddbc92bb81e41fa8404105448d")

	tests := []struct {
		name     string
		programs []solanacarrier.AllowedProgram
		contains string
	}{
		{name: "nil list", programs: nil},
		{name: "empty list", programs: []solanacarrier.AllowedProgram{}},
		{name: "cap of one", programs: distinctPrograms(1, 1)},
		{name: "cap of 64", programs: distinctPrograms(1, 64)},
		{name: "16 entries", programs: distinctPrograms(16, 4)},
		{
			name:     "17 entries",
			programs: distinctPrograms(17, 4),
			contains: "at most 16 programs may be allowed, got 17",
		},
		{
			name:     "31-byte program id",
			programs: []solanacarrier.AllowedProgram{{ProgramId: bytes.Repeat([]byte{0x20}, 31), MaxInstructions: 4}},
			contains: "entry 0: program id must be 32 bytes, got 31",
		},
		{
			name:     "33-byte program id",
			programs: []solanacarrier.AllowedProgram{{ProgramId: bytes.Repeat([]byte{0x20}, 33), MaxInstructions: 4}},
			contains: "entry 0: program id must be 32 bytes, got 33",
		},
		{
			name:     "missing program id",
			programs: []solanacarrier.AllowedProgram{{MaxInstructions: 4}},
			contains: "entry 0: program id must be 32 bytes, got 0",
		},
		{
			name: "duplicate program id",
			programs: []solanacarrier.AllowedProgram{
				{ProgramId: lighthouseProgramKey, MaxInstructions: 4},
				{ProgramId: computeBudgetProgramKey, MaxInstructions: 4},
				{ProgramId: lighthouseProgramKey, MaxInstructions: 1},
			},
			contains: "entry 2: program L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95 is listed more than once",
		},
		{
			name:     "memo program",
			programs: []solanacarrier.AllowedProgram{{ProgramId: memoProgram, MaxInstructions: 1}},
			contains: "entry 0: the SPL Memo program is always bound",
		},
		{
			name:     "cap of zero",
			programs: []solanacarrier.AllowedProgram{{ProgramId: computeBudgetProgramKey, MaxInstructions: 0}},
			contains: "entry 0: program ComputeBudget111111111111111111111111111111 max instructions must be between 1 and 64, got 0",
		},
		{
			name:     "cap of 65",
			programs: []solanacarrier.AllowedProgram{{ProgramId: computeBudgetProgramKey, MaxInstructions: 65}},
			contains: "entry 0: program ComputeBudget111111111111111111111111111111 max instructions must be between 1 and 64, got 65",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := solanacarrier.DefaultParams()
			params.AllowedPrograms = tc.programs

			err := params.Validate()
			if tc.contains == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, solanacarrier.ErrInvalidAllowedPrograms)
			require.ErrorContains(t, err, tc.contains)
		})
	}
}
