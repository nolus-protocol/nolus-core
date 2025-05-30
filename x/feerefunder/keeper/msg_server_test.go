package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/testutil"
	"github.com/Nolus-Protocol/nolus-core/testutil/feerefunder/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
)

func TestMsgUpdateParamsValidate(t *testing.T) {
	k, ctx := keeper.FeeKeeper(t, nil, nil)

	tests := []struct {
		name        string
		msg         types.MsgUpdateParams
		expectedErr string
	}{
		{
			"empty authority",
			types.MsgUpdateParams{
				Authority: "",
			},
			"authority is invalid",
		},
		{
			"invalid authority",
			types.MsgUpdateParams{
				Authority: "invalid authority",
			},
			"authority is invalid",
		},
		{
			"invalid ack fee",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					MinFee: types.Fee{
						RecvFee: nil,
						AckFee: sdk.Coins{
							{
								Denom:  "{}!@#a",
								Amount: math.NewInt(100),
							},
						},
						TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(100))),
					},
				},
			},
			sdkerrors.ErrInvalidCoins.Error(),
		},
		{
			"invalid timeout fee",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					MinFee: types.Fee{
						RecvFee: nil,
						AckFee:  sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(100))),
						TimeoutFee: sdk.Coins{
							{
								Denom:  params.DefaultBondDenom,
								Amount: math.NewInt(-100),
							},
						},
					},
				},
			},
			sdkerrors.ErrInvalidCoins.Error(),
		},
		{
			"non-zero recv fee",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					MinFee: types.Fee{
						RecvFee:    sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(100))),
						AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(100))),
						TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(100))),
					},
				},
			},
			sdkerrors.ErrInvalidCoins.Error(),
		},
		{
			"zero ack fee",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					MinFee: types.Fee{
						RecvFee:    nil,
						AckFee:     nil,
						TimeoutFee: sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(100))),
					},
				},
			},
			sdkerrors.ErrInvalidCoins.Error(),
		},
		{
			"zero timeout fee",
			types.MsgUpdateParams{
				Authority: testutil.TestOwnerAddress,
				Params: types.Params{
					MinFee: types.Fee{
						RecvFee:    nil,
						AckFee:     sdk.NewCoins(sdk.NewCoin(params.DefaultBondDenom, math.NewInt(100))),
						TimeoutFee: nil,
					},
				},
			},
			sdkerrors.ErrInvalidCoins.Error(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resp, err := k.UpdateParams(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}
