package simulation_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/simulation"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)
	tmAddr := ed25519.GenPrivKey().PubKey().Address().String()
	address, _ := sdk.AccAddressFromHex(tmAddr)

	suspendedState := types.NewSuspendedState(address.String(), true, 1245)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.SuspendStateKey, Value: cdc.MustMarshal(&suspendedState)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Minter", fmt.Sprintf("%v\n%v", suspendedState, suspendedState)},
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
