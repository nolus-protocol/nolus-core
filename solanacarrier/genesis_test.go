package solanacarrier_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

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

func TestGenesisValidateRejectsBadAllowlist(t *testing.T) {
	params := solanacarrier.DefaultParams()
	params.AllowedPrograms[1].MaxInstructions = 0

	gs := solanacarrier.GenesisState{Params: params}
	require.ErrorIs(t, gs.Validate(), solanacarrier.ErrInvalidAllowedPrograms)
}

func TestDefaultGenesisRoundTripsThroughJSON(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	bz, err := cdc.MarshalJSON(solanacarrier.DefaultGenesis())
	require.NoError(t, err)
	require.Contains(t, string(bz), `"allowed_programs":[`)
	require.Contains(t, string(bz), `"program_id":"`+base64.StdEncoding.EncodeToString(lighthouseProgramKey)+`"`)
	require.Contains(t, string(bz), `"max_instructions":4`)

	var decoded solanacarrier.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(bz, &decoded))
	require.Equal(t, *solanacarrier.DefaultGenesis(), decoded)
}
