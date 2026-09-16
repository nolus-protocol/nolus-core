package solanacarrier_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

func TestDefaultGenesisValidates(t *testing.T) {
	gs := solanacarrier.DefaultGenesis()
	require.NoError(t, gs.Validate())
	require.Equal(t, solanacarrier.DefaultParams(), gs.Params)
}

func TestGenesisValidateRejectsBadParams(t *testing.T) {
	gs := solanacarrier.GenesisState{Params: solanacarrier.Params{FeePayer: []byte{0x01}}}
	require.ErrorIs(t, gs.Validate(), solanacarrier.ErrInvalidFeePayer)
}
