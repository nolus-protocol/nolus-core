package simulation

import (
	"bytes"
	"fmt"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding suspended type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key, types.SuspendStateKey):
			var suspendedStateA, suspendedStateB types.SuspendedState
			cdc.MustUnmarshal(kvA.Value, &suspendedStateA)
			cdc.MustUnmarshal(kvB.Value, &suspendedStateB)
			return fmt.Sprintf("%v\n%v", suspendedStateA, suspendedStateB)
		default:
			panic(fmt.Sprintf("invalid suspended key %X", kvA.Key))
		}
	}
}
