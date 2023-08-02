package simulation_test

// refactor: fix when simulation refactor is done
// import (
// 	"math/rand"
// 	"testing"

// 	"github.com/Nolus-Protocol/nolus-core/x/tax/simulation"
// 	"github.com/stretchr/testify/require"
// )

// func TestParamChanges(t *testing.T) {
// 	s := rand.NewSource(1)
// 	r := rand.New(s)

// 	expected := []struct {
// 		composedKey string
// 		key         string
// 		simValue    string
// 		subspace    string
// 	}{
// 		{"tax/FeeRate", "FeeRate", "35", "tax"},
// 	}

// 	paramChanges := simulation.ParamChanges(r)
// 	require.Len(t, paramChanges, 1)

// 	for i, p := range paramChanges {
// 		require.Equal(t, expected[i].composedKey, p.ComposedKey())
// 		require.Equal(t, expected[i].key, p.Key())
// 		require.Equal(t, expected[i].simValue, p.SimValue()(r))
// 		require.Equal(t, expected[i].subspace, p.Subspace())
// 	}
// }
