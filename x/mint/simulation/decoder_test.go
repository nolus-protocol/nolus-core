package simulation_test

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/Nolus-Protocol/nolus-core/custom/util"
// 	"github.com/cosmos/cosmos-sdk/testutil/sims"
// 	sdk "github.com/cosmos/cosmos-sdk/types"

// 	"github.com/stretchr/testify/require"

// 	"github.com/Nolus-Protocol/nolus-core/x/mint/simulation"
// 	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
// 	"github.com/cosmos/cosmos-sdk/types/kv"
// )

// refactor: fix when simulation refactor is done
// func TestDecodeStore(t *testing.T) {
// 	cdc := sims.MakeTestEncodingConfig().Marshaler
// 	dec := simulation.NewDecodeStore(cdc)

// 	minter := types.NewMinter(sdk.MustNewDecFromStr("13.123456789"), sdk.NewUint(10003145), sdk.NewUint(uint64(util.GetCurrentTimeUnixNano())), sdk.ZeroUint())

// 	kvPairs := kv.Pairs{
// 		Pairs: []kv.Pair{
// 			{Key: types.MinterKey, Value: cdc.MustMarshal(&minter)},
// 			{Key: []byte{0x99}, Value: []byte{0x99}},
// 		},
// 	}
// 	tests := []struct {
// 		name        string
// 		expectedLog string
// 	}{
// 		{"Minter", fmt.Sprintf("%v\n%v", minter, minter)},
// 		{"other", ""},
// 	}

// 	for i, tt := range tests {
// 		i, tt := i, tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			switch i {
// 			case len(tests) - 1:
// 				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
// 			default:
// 				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
// 			}
// 		})
// 	}
// }
