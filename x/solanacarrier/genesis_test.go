package solanacarrier_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/keeper"
	solanacarrier "github.com/Nolus-Protocol/nolus-core/x/solanacarrier"
	"github.com/Nolus-Protocol/nolus-core/x/solanacarrier/types"
)

func TestInitExportGenesisRoundTrip(t *testing.T) {
	k, ctx := testkeeper.SolanaCarrierKeeper(t, nil)

	genState := types.DefaultGenesis()
	solanacarrier.InitGenesis(ctx, *k, *genState)

	require.Equal(t, genState, solanacarrier.ExportGenesis(ctx, *k))
}

func TestInitGenesisPanicsOnInvalidParams(t *testing.T) {
	k, ctx := testkeeper.SolanaCarrierKeeper(t, nil)

	require.Panics(t, func() {
		solanacarrier.InitGenesis(ctx, *k, types.GenesisState{Params: types.Params{FeePayer: []byte{0x01}}})
	})
}

func TestExportGenesisPanicsWhenParamsUnset(t *testing.T) {
	k, ctx := testkeeper.SolanaCarrierKeeper(t, nil)

	require.Panics(t, func() {
		solanacarrier.ExportGenesis(ctx, *k)
	})
}
