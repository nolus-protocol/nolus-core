package keeper_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/keeper"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"
)

func (suite *KeeperTestSuite) TestTaxDecorator() {
	const rnDenom = "atom"

	baseDenom, err := sdk.GetBaseDenom()
	suite.Require().NoError(err)

	testCases := []struct {
		title              string
		feeDenoms          []string
		feeAmount          sdk.Int
		expPass            bool
		expTreasuryBalance sdk.Coins
		expErr             error
	}{
		{
			title:              "successful tax deduction should increase the treasury balance",
			feeDenoms:          []string{baseDenom},
			feeAmount:          sdk.NewInt(10),
			expPass:            true,
			expTreasuryBalance: sdk.NewCoins(sdk.NewCoin(baseDenom, sdk.NewInt(4))),
			expErr:             nil,
		},
		{
			title:              "tx without fees should continue to the next AnteHandler",
			feeDenoms:          []string{},
			feeAmount:          sdk.NewInt(0),
			expPass:            true,
			expTreasuryBalance: sdk.NewCoins(),
			expErr:             nil,
		},
		{
			title:              "pay fees with insufficient funds should fail",
			feeDenoms:          []string{baseDenom},
			feeAmount:          sdk.NewInt(100000),
			expPass:            false,
			expTreasuryBalance: nil,
			expErr:             sdkerrors.ErrInsufficientFunds,
		},
		{
			title:              "pay fees with not allowed denom should fail",
			feeDenoms:          []string{rnDenom},
			feeAmount:          sdk.NewInt(100),
			expPass:            false,
			expTreasuryBalance: nil,
			expErr:             types.ErrInvalidFeeDenom,
		},
		{
			title:              "pay fees with multiple denoms should fail",
			feeDenoms:          []string{baseDenom, rnDenom},
			feeAmount:          sdk.NewInt(100),
			expPass:            false,
			expTreasuryBalance: nil,
			expErr:             types.ErrTooManyFeeCoins,
		},
	}

	for _, tc := range testCases {
		// reset pool and accounts for each test
		suite.SetupTest(true)

		suite.Run(tc.title, func() {
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

			// create account with nolus and atom
			accs := suite.CreateTestAccounts(1)
			addr := accs[0].acc.GetAddress()
			priv := accs[0].priv

			var coins sdk.Coins
			coins = coins.Add(sdk.NewInt64Coin(baseDenom, 500))
			coins = coins.Add(sdk.NewInt64Coin(rnDenom, 300))
			suite.FundAcc(addr, coins)

			// set gas
			gasLimit := sdktestutil.NewTestGasLimit()
			suite.txBuilder.SetGasLimit(gasLimit)

			// msg and signatures
			msg := sdktestutil.NewTestMsg(addr)
			suite.Require().NoError(suite.txBuilder.SetMsgs(msg))

			// create tx
			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv}, []uint64{0}, []uint64{0}
			tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
			suite.Require().NoError(err)

			// set account
			suite.app.AccountKeeper.SetAccount(suite.ctx, accs[0].acc)

			// get chained ante handler
			dfd := ante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, nil)
			dtd := keeper.NewDeductTaxDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.TaxKeeper)
			anteHandler := sdk.ChainAnteDecorators(dfd, dtd)

			// retrieve treasury address
			treasuryAddr, err := sdk.AccAddressFromBech32(suite.app.TaxKeeper.ContractAddress(suite.ctx))
			suite.Require().NoError(err)

			// add coins to pay the tax
			var txFees sdk.Coins
			for _, feeDenom := range tc.feeDenoms {
				txFees = txFees.Add(sdk.NewCoin(feeDenom, tc.feeAmount))
			}

			suite.txBuilder.SetFeeAmount(txFees)

			_, err = anteHandler(suite.ctx, tx, false)

			if tc.expPass {
				suite.Require().NoError(err, "test: %s", tc.title)

				treasuryBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, treasuryAddr)
				suite.Require().Equal(tc.expTreasuryBalance, treasuryBalance, "Treasury should have collected correct tax amount")
			} else {
				suite.Require().Error(err, "test: %s", tc.title)
				suite.ErrorIs(err, tc.expErr, tc.title)
			}
		})
	}
}
