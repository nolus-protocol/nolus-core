package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/Nolus-Protocol/nolus-core/x/mint/simulation"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abnormal scenarios are not tested here.
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
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var mintGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &mintGenesis)

	require.Equal(t, sdkmath.NewUint(uint64(time.Second.Nanoseconds()*13)), mintGenesis.Params.MaxMintableNanoseconds)
	require.Equal(t, "0", mintGenesis.Minter.TotalMinted.String())
	require.Equal(t, "0.000000000000000000", mintGenesis.Minter.NormTimePassed.String())
	require.Equal(t, sdkmath.ZeroUint(), mintGenesis.Minter.PrevBlockTimestamp)
}

// TestRandomizedGenState tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState1(t *testing.T) {
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

// TestGenMaxMintableNanoseconds tests for generation of MaxMintableNanoseconds with different given rand sources.
func TestGenMaxMintableNanoseconds(t *testing.T) {
	tests := []struct {
		r                   *rand.Rand
		expectedMaxMintable sdkmath.Uint
	}{
		{rand.New(rand.NewSource(1)), sdkmath.NewUint(uint64((4000000000)))},
		{rand.New(rand.NewSource(0)), sdkmath.NewUint(uint64((50000000000)))},
		{rand.New(rand.NewSource(1241255)), sdkmath.NewUint(uint64((13000000000)))},
		{rand.New(rand.NewSource(4)), sdkmath.NewUint(uint64((21000000000)))},
		{rand.New(rand.NewSource(17)), sdkmath.NewUint(uint64((12000000000)))},
		{rand.New(rand.NewSource(60)), sdkmath.NewUint(uint64((35000000000)))},
		{rand.New(rand.NewSource(22)), sdkmath.NewUint(uint64((42000000000)))},
		{rand.New(rand.NewSource(-2)), sdkmath.NewUint(uint64((25000000000)))},
	}

	for _, tt := range tests {
		actualMaxMintable := simulation.GenMaxMintableNanoseconds(tt.r)
		require.Equal(t, tt.expectedMaxMintable, actualMaxMintable)
	}
}
