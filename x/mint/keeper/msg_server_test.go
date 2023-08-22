package keeper_test

import (
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		request   *types.MsgUpdateParams
		expectErr bool
	}{
		{
			name: "set invalid authority",
			request: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
		},
		{
			name: "set invalid params",
			request: &types.MsgUpdateParams{
				Authority: s.app.MintKeeper.GetAuthority(),
				Params: types.Params{
					MintDenom:              sdk.DefaultBondDenom,
					MaxMintableNanoseconds: sdk.NewUint(0), // invalid
				},
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: s.app.MintKeeper.GetAuthority(),
				Params: types.Params{
					MintDenom:              sdk.DefaultBondDenom,
					MaxMintableNanoseconds: sdk.NewUint(60000000000), // 1 min default
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.request)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
