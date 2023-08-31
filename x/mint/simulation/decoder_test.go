package simulation_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/kv"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/Nolus-Protocol/nolus-core/custom/util"
	"github.com/Nolus-Protocol/nolus-core/x/mint/simulation"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	dec := simulation.NewDecodeStore(cdc)

	minter := types.NewMinter(sdkmath.LegacyMustNewDecFromStr("13.123456789"), sdkmath.NewUint(10003145), sdkmath.NewUint(uint64(util.GetCurrentTimeUnixNano())), sdkmath.ZeroUint())

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.MinterKey, Value: cdc.MustMarshal(&minter)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Minter", fmt.Sprintf("%v\n%v", minter, minter)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
