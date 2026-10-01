package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/x/solanacarrier/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	gs := types.DefaultGenesis()
	require.NoError(t, gs.Validate())
	require.Equal(t, types.DefaultParams(), gs.Params)
}

func TestGenesisValidateRejectsBadParams(t *testing.T) {
	gs := types.GenesisState{Params: types.Params{FeePayer: []byte{0x01}}}
	require.ErrorIs(t, gs.Validate(), types.ErrInvalidFeePayer)
}
