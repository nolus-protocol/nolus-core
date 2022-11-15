package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/simulation"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abonormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: 1000,
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var taxGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &taxGenesis)

	require.Equal(t, "stake", taxGenesis.Params.BaseDenom)
	require.GreaterOrEqual(t, taxGenesis.Params.FeeRate, int32(1))
	require.GreaterOrEqual(t, int32(100), taxGenesis.Params.FeeRate)
	require.Equal(t, "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz", taxGenesis.Params.ContractAddress)
}

// TestRandomizedGenState tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenStateAbnormal(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)
	// all these tests will panic
	tests := []struct {
		simState module.SimulationState
		panicMsg string
	}{
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       cdc,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}

// TestGenRandomFeeRate tests for generation of FeeRate with different given rand sources.
func TestGenRandomFeeRate(t *testing.T) {
	tests := []struct {
		r               *rand.Rand
		expectedFeeRate int32
	}{
		{rand.New(rand.NewSource(1)), int32(24)},
		{rand.New(rand.NewSource(0)), int32(37)},
		{rand.New(rand.NewSource(1241255)), int32(78)},
		{rand.New(rand.NewSource(4)), int32(89)},
		{rand.New(rand.NewSource(17)), int32(4)},
		{rand.New(rand.NewSource(60)), int32(16)},
		{rand.New(rand.NewSource(22)), int32(55)},
		{rand.New(rand.NewSource(-2)), int32(47)},
	}

	for _, tt := range tests {
		actualFeeRate := simulation.GenRandomFeeRate(tt.r)
		require.Equal(t, tt.expectedFeeRate, actualFeeRate)
	}
}
