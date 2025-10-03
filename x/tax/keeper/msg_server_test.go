package keeper_test

import (
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
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
				Authority: s.app.TaxKeeper.GetAuthority(),
				Params: types.Params{
					FeeRate:         0,
					TreasuryAddress: "",
					BaseDenom:       "",
				},
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: s.app.TaxKeeper.GetAuthority(),
				Params: types.Params{
					FeeRate:         1,
					TreasuryAddress: "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz",
					BaseDenom:       "nolus",
				},
			},
			expectErr: false,
		},
		{
			name: "set full valid params fee rate 100",
			request: &types.MsgUpdateParams{
				Authority: s.app.TaxKeeper.GetAuthority(),
				Params: types.Params{
					FeeRate:         100,
					TreasuryAddress: "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz",
					BaseDenom:       "unls",
				},
			},
			expectErr: false,
		},
		{
			name: "set full valid params fee rate 0",
			request: &types.MsgUpdateParams{
				Authority: s.app.TaxKeeper.GetAuthority(),
				Params: types.Params{
					FeeRate:         0,
					TreasuryAddress: "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz",
					BaseDenom:       "nolus",
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.SetupTest(false)

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
