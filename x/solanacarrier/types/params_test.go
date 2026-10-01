package types_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/x/solanacarrier/types"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func TestParamsValidateDefault(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())
	require.Len(t, types.DefaultFeePayer, 32)
}

func TestParamsValidateWrongLength(t *testing.T) {
	for name, feePayer := range map[string][]byte{
		"nil":       nil,
		"empty":     {},
		"31 bytes":  types.DefaultFeePayer[:31],
		"33 bytes":  append(append([]byte{}, types.DefaultFeePayer...), 0x00),
		"truncated": types.DefaultFeePayer[:16],
	} {
		t.Run(name, func(t *testing.T) {
			err := types.Params{FeePayer: feePayer}.Validate()
			require.ErrorIs(t, err, types.ErrInvalidFeePayer)
			require.Contains(t, err.Error(), "32 bytes")
		})
	}
}

func TestParamsValidateRejectsOnCurveFeePayer(t *testing.T) {
	// The SPL Memo program id is a plain on-curve ed25519 point.
	onCurve := mustHex(t, "054a535a992921064d24e87160da387c7c35b5ddbc92bb81e41fa8404105448d")

	err := types.Params{FeePayer: onCurve}.Validate()
	require.ErrorIs(t, err, types.ErrInvalidFeePayer)
	require.True(t, strings.Contains(err.Error(), "on-curve"), "error must name on-curve, got %q", err)
}

func TestDefaultFeePayerIsTheSentinelPDA(t *testing.T) {
	require.Equal(t,
		mustHex(t, "cefc037c11436dc85b6943b43c94ec50c0493696cb7e797e84c1763038e9ad70"),
		types.DefaultFeePayer,
	)
}
