package keeper_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/Nolus-Protocol/nolus-core/app"
// 	"github.com/Nolus-Protocol/nolus-core/testutil/nullify"
// 	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
// 	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
// 	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
// 	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
// 	"github.com/stretchr/testify/suite"
// )

// // TestAccount represents a client Account that can be used in unit tests.
// type TestAccount struct {
// 	acc  authtypes.AccountI
// 	priv cryptotypes.PrivKey
// }

// // KeeperTestSuite is a test suite to be used with mint's keeper tests.
// type KeeperTestSuite struct {
// 	suite.Suite

// 	app           *app.App
// 	ctx           sdk.Context
// 	sdkWrappedCtx context.Context
// }

// func (s *KeeperTestSuite) TestSetMinter() {
// 	s.SetupTest(false)
// 	minterKeeper := s.app.MintKeeper

// 	got := minterKeeper.GetMinter(s.ctx)
// 	s.Require().NotNil(got)

// 	nullify.Fill(got)
// }

// func (s *KeeperTestSuite) TestGetParamsWithDefaultParams() {
// 	s.SetupTest(false)
// 	minterKeeper := s.app.MintKeeper

// 	got := minterKeeper.GetParams(s.ctx)
// 	s.Require().NotNil(got)
// 	s.Require().Equal(types.DefaultParams(), got)
// }

// func (s *KeeperTestSuite) TestMintCoins() {
// 	s.SetupTest(false)
// 	minterKeeper := s.app.MintKeeper

// 	err := minterKeeper.MintCoins(s.ctx, sdk.Coins{sdk.Coin{Denom: "nolus", Amount: sdk.NewInt(200)}})
// 	s.Require().Nil(err)
// }

// func (s *KeeperTestSuite) TestMintCoinsNoCoinsPassed() {
// 	s.SetupTest(false)
// 	minterKeeper := s.app.MintKeeper

// 	err := minterKeeper.MintCoins(s.ctx, sdk.Coins{})
// 	s.Require().Nil(err)
// }

// func (s *KeeperTestSuite) TestAddCollectedFees() {
// 	s.SetupTest(false)
// 	minterKeeper := s.app.MintKeeper

// 	baseDenom := s.app.TaxKeeper.BaseDenom(s.ctx)
// 	// create account and fund it with 5000 of base denom
// 	accs := s.createTestAccounts(1)
// 	addr := accs[0].acc.GetAddress()
// 	coins := sdk.Coins{sdk.NewInt64Coin(baseDenom, 5000)}
// 	s.fundAcc(addr, coins)

// 	// fund mint's account so it can execute AddCollectedFees successfully
// 	err := s.app.BankKeeper.SendCoins(s.ctx, addr, s.app.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.NewCoins(sdk.NewCoin(baseDenom, sdk.NewInt(int64(1000)))))
// 	s.Require().Nil(err)

// 	err = minterKeeper.AddCollectedFees(s.ctx, sdk.Coins{sdk.NewCoin(baseDenom, sdk.NewInt(int64(200)))})
// 	s.Require().Nil(err)

// 	feeCollectorBalance := s.app.BankKeeper.GetBalance(s.ctx, s.app.AccountKeeper.GetModuleAddress("fee_collector"), baseDenom)
// 	s.Require().Equal(int64(200), feeCollectorBalance.Amount.Int64())
// }

// // createTestAccounts creates accounts.
// func (s *KeeperTestSuite) createTestAccounts(numAccs int) []TestAccount {
// 	var accounts []TestAccount
// 	for i := 0; i < numAccs; i++ {
// 		priv, _, addr := sdktestutil.KeyTestPubAddr()
// 		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
// 		s.Require().NoError(acc.SetAccountNumber(uint64(i)))
// 		s.app.AccountKeeper.SetAccount(s.ctx, acc)
// 		accounts = append(accounts, TestAccount{acc, priv})
// 	}

// 	return accounts
// }

// // fundAcc funds target address with specified amount.
// func (s *KeeperTestSuite) fundAcc(addr sdk.AccAddress, amounts sdk.Coins) {
// 	err := banktestutil.FundAccount(s.app.BankKeeper, s.ctx, addr, amounts)
// 	s.Require().NoError(err)
// }

// func TestAnteTestSuite(t *testing.T) {
// 	suite.Run(t, new(KeeperTestSuite))
// }
