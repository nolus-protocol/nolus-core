package typesv2_test

import (
	"testing"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	params.GetDefaultConfig()
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc:     "valid genesis state",
			genState: &types.GenesisState{Params: types.NewParams(types.DefaultFeeRate, types.DefaultTreasuryAddress, types.DefaultBaseDenom)},
			valid:    true,
		},
		{
			desc:     "invalid genesis state",
			genState: &types.GenesisState{},
			valid:    false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
