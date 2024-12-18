package keeper_test

import (
	"testing"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	testkeeper "github.com/Nolus-Protocol/nolus-core/testutil/keeper"
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func (s *KeeperTestSuite) TestParams() {
	testCases := []struct {
		name      string
		input     types.Params
		expectErr bool
	}{
		{
			name: "set invalid params",
			input: types.Params{
				FeeRate:         0,
				TreasuryAddress: "a",
				BaseDenom:       "1",
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			input: types.Params{
				FeeRate:         1,
				TreasuryAddress: "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz",
				BaseDenom:       "nolus",
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.SetupTest(false)

		s.Run(tc.name, func() {
			expected, err := s.app.TaxKeeper.GetParams(s.ctx)
			s.Require().NoError(err)
			// TODO expect panic if params are not set
			err = s.app.TaxKeeper.SetParams(s.ctx, tc.input)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				expected = tc.input
				s.Require().NoError(err)
			}

			p, err := s.app.TaxKeeper.GetParams(s.ctx)
			s.Require().NoError(err)
			s.Require().Equal(expected, p)
		})
	}
}

func TestGetParams(t *testing.T) {
	params.GetDefaultConfig()
	k, ctx := testkeeper.TaxKeeper(t, false, sdk.DecCoins{}, types.DefaultParams())
	params := types.DefaultParams()

	actualParams, err := k.GetParams(ctx)
	require.NoError(t, err)

	require.EqualValues(t, params, actualParams)
	require.EqualValues(t, params.FeeRate, k.FeeRate(ctx))
	require.EqualValues(t, params.TreasuryAddress, k.TreasuryAddress(ctx))
	require.EqualValues(t, params.BaseDenom, k.BaseDenom(ctx))
}
