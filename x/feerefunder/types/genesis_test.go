package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
)

const (
	TestAddressNolus         = "nolus10pslfpsx0l3rt0cm8rhmgg75exnt3ruqrprs28"
	TestContractAddressJuno  = "juno10h0hc64jv006rr8qy0zhlu4jsxct8qwa0vtaleayh0ujz0zynf2s2r7v8q"
	TestContractAddressNolus = "nolus1f6cu6ypvpyh0p8d7pqnps2pduj87hda5t9v4mqrc8ra67xp28uwq4f4ysz"
)

func TestGenesisState_Validate(t *testing.T) {
	cfg := params.GetDefaultConfig()
	cfg.Seal()

	validRecvFee := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(0)))
	validAckFee := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(types.DefaultFees.AckFee.AmountOf(params.DefaultBondDenom).Int64()+1)))
	validTimeoutFee := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(types.DefaultFees.TimeoutFee.AmountOf(params.DefaultBondDenom).Int64()+1)))

	invalidRecvFee := sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(1)))

	validPacketID := types.NewPacketID("port", "channel-1", 64)

	for _, tc := range []struct {
		desc             string
		genState         *types.GenesisState
		valid            bool
		expectedErrorMsg string
	}{
		{
			desc:             "default is valid",
			genState:         types.DefaultGenesis(),
			valid:            true,
			expectedErrorMsg: "",
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNolus,
					PacketId: validPacketID,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid:            true,
			expectedErrorMsg: "",
		},
		{
			desc: "invalid payer address",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    "address",
					PacketId: validPacketID,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid:            false,
			expectedErrorMsg: "failed to parse the payer address",
		},
		{
			desc: "payer is not a contract",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestAddressNolus,
					PacketId: validPacketID,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid:            false,
			expectedErrorMsg: "is not a contract",
		},
		{
			desc: "payer is from a wrong chain",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressJuno,
					PacketId: validPacketID,
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid:            false,
			expectedErrorMsg: "failed to parse the payer address",
		},
		{
			desc: "invalid port",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNolus,
					PacketId: types.NewPacketID("*", "channel", 64),
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid:            false,
			expectedErrorMsg: "port id",
		},
		{
			desc: "invalid channel",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNolus,
					PacketId: types.NewPacketID("port", "*", 64),
					Fee: types.Fee{
						RecvFee:    validRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid:            false,
			expectedErrorMsg: "channel id",
		},
		{
			desc: "Recv fee non-zero",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNolus,
					PacketId: validPacketID,
					Fee: types.Fee{
						RecvFee:    invalidRecvFee,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid:            false,
			expectedErrorMsg: "invalid fees",
		},
		{
			desc: "Recv fee nil",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				FeeInfos: []types.FeeInfo{{
					Payer:    TestContractAddressNolus,
					PacketId: validPacketID,
					Fee: types.Fee{
						RecvFee:    nil,
						AckFee:     validAckFee,
						TimeoutFee: validTimeoutFee,
					},
				}},
			},
			valid: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrorMsg)
			}
		})
	}
}
