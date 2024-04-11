package keeper_test

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/testutil/nullify"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxtypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
)

// TestAccount represents a client Account that can be used in unit tests.
type TestAccount struct {
	acc  sdk.AccountI
	priv cryptotypes.PrivKey
}

// KeeperTestSuite is a test suite to be used with mint's keeper tests.
type KeeperTestSuite struct {
	suite.Suite

	app       *app.App
	ctx       context.Context
	msgServer types.MsgServer
}

func (s *KeeperTestSuite) TestParams() {
	testCases := []struct {
		name      string
		input     types.Params
		expectErr bool
	}{
		{
			name: "set valid params",
			input: types.Params{
				MintDenom:              "nolus",
				MaxMintableNanoseconds: sdkmath.NewUint(60000000000), // 1 min default
			},
			expectErr: false,
		},
		{
			name: "set invalid params",
			input: types.Params{
				MintDenom:              "error-denom",
				MaxMintableNanoseconds: sdkmath.NewUint(0), // 0 invalid
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.SetupTest(false)

		s.Run(tc.name, func() {
			expected := s.app.MintKeeper.GetParams(s.ctx)
			err := s.app.MintKeeper.SetParams(s.ctx, tc.input)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				expected = tc.input
				s.Require().NoError(err)
			}

			p := s.app.MintKeeper.GetParams(s.ctx)
			s.Require().Equal(expected, p)
		})
	}
}

func (s *KeeperTestSuite) TestSetMinter() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	got := minterKeeper.GetMinter(s.ctx)
	s.Require().NotNil(got)

	nullify.Fill(got)
}

func (s *KeeperTestSuite) TestGetParamsWithDefaultParams() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	got := minterKeeper.GetParams(s.ctx)
	s.Require().NotNil(got)
	s.Require().Equal(types.DefaultParams(), got)
}

func (s *KeeperTestSuite) TestMintCoins() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	err := minterKeeper.MintCoins(s.ctx, sdk.Coins{sdk.Coin{Denom: "nolus", Amount: sdkmath.NewInt(200)}})
	s.Require().Nil(err)
}

func (s *KeeperTestSuite) TestMintCoinsNoCoinsPassed() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	err := minterKeeper.MintCoins(s.ctx, sdk.Coins{})
	s.Require().Nil(err)
}

func (s *KeeperTestSuite) TestAddCollectedFees() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper
	sdkctx := sdk.UnwrapSDKContext(s.ctx)

	_ = s.app.TaxKeeper.SetParams(sdkctx, taxtypes.DefaultParams())
	taxParams := s.app.TaxKeeper.GetParams(sdkctx)
	// create account and fund it with 5000 of base denom
	accs := s.createTestAccounts(1)
	addr := accs[0].acc.GetAddress()
	coins := sdk.Coins{sdk.NewInt64Coin(taxParams.BaseDenom, 5000)}
	s.fundAcc(addr, coins)

	// fund mint's account so it can execute AddCollectedFees successfully
	err := s.app.BankKeeper.SendCoins(s.ctx, addr, s.app.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.NewCoins(sdk.NewCoin(taxParams.BaseDenom, sdkmath.NewInt(int64(1000)))))
	s.Require().Nil(err)

	err = minterKeeper.AddCollectedFees(s.ctx, sdk.Coins{sdk.NewCoin(taxParams.BaseDenom, sdkmath.NewInt(int64(200)))})
	s.Require().Nil(err)

	feeCollectorBalance := s.app.BankKeeper.GetBalance(s.ctx, s.app.AccountKeeper.GetModuleAddress("fee_collector"), taxParams.BaseDenom)
	s.Require().Equal(int64(200), feeCollectorBalance.Amount.Int64())
}

// createTestAccounts creates accounts.
func (s *KeeperTestSuite) createTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount
	for i := 0; i < numAccs; i++ {
		priv, _, addr := sdktestutil.KeyTestPubAddr()
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
		s.Require().NoError(acc.SetAccountNumber(uint64(i)))
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// fundAcc funds target address with specified amount.
func (s *KeeperTestSuite) fundAcc(addr sdk.AccAddress, amounts sdk.Coins) {
	err := banktestutil.FundAccount(s.ctx, s.app.BankKeeper, addr, amounts)
	s.Require().NoError(err)
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
