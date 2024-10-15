package feerefunder_test

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/testutil/feerefunder/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/testutil/common/nullify"
	"github.com/Nolus-Protocol/nolus-core/x/feerefunder"
	"github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
)

const TestContractAddressNeutron = "neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq"

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		FeeInfos: []types.FeeInfo{{
			Payer:    TestContractAddressNeutron,
			PacketId: types.NewPacketID("port", "channel-1", 64),
			Fee: types.Fee{
				RecvFee:    sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(0))),
				AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(types.DefaultFees.AckFee.AmountOf(params.DefaultBondDenom).Int64()+1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(types.DefaultFees.TimeoutFee.AmountOf(params.DefaultBondDenom).Int64()+1))),
			},
		}},
	}

	require.EqualValues(t, genesisState.Params, types.DefaultParams())

	k, ctx := keeper.FeeKeeper(t, nil, nil)
	feerefunder.InitGenesis(ctx, *k, genesisState)
	got := feerefunder.ExportGenesis(ctx, *k)

	require.EqualValues(t, got.Params, types.DefaultParams())
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
