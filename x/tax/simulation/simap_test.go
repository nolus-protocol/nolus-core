package simulation_test

import (
	"math/rand"
	"testing"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/simulation"
)

func TestFindAccount(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	randomAccounts := simtypes.RandomAccounts(r, 5)

	account, isAccInArray := simulation.FindAccount(randomAccounts, randomAccounts[2].Address.String())
	require.True(t, isAccInArray)
	require.Equal(t, account.Address.String(), randomAccounts[2].Address.String())
}

func TestFindAccountPanic(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	randomAccounts := simtypes.RandomAccounts(r, 5)

	require.Panics(t, assert.PanicTestFunc(func() {
		simulation.FindAccount(randomAccounts, "not-existing")
	}))
}
