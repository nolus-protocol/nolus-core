package types_test

import (
	types2 "gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
    for _, tc := range []struct {
    		desc          string
    		genState      *types2.GenesisState
    		valid bool
    } {
        {
            desc:     "default is valid",
            genState: types2.DefaultGenesis(),
            valid:    true,
        },
        {
            desc:     "valid genesis state",
            genState: &types2.GenesisState{
                // this line is used by starport scaffolding # types/genesis/validField
            },
            valid:    true,
        },
        // this line is used by starport scaffolding # types/genesis/testcase
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
