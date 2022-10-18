package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/app/params"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/types"
)

func TestGenesisState_Validate(t *testing.T) {
	params.SetAddressPrefixes()
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
			genState: &types.GenesisState{Params: types.NewParams(100, "nolus", types.DefaultContractAddress)},
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
