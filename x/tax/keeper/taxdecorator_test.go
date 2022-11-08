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

// func (suite *KeeperTestSuite) TestTaxes() {
// 	suite.SetupTest(true)
// 	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

// 	// keys and addresses
// 	priv1, _, addr1 := sdktestutil.KeyTestPubAddr()

// 	// msg and signatures
// 	msg := sdktestutil.NewTestMsg(addr1)

// 	var feeAmount sdk.Coins
// 	feeAmount = feeAmount.Add(sdk.NewCoin(NOLUS_DENOM, sdk.NewInt(150)))

// 	gasLimit := sdktestutil.NewTestGasLimit()
// 	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
// 	suite.txBuilder.SetFeeAmount(feeAmount)
// 	suite.txBuilder.SetGasLimit(gasLimit)

// 	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
// 	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
// 	suite.Require().NoError(err)

// 	// Set account with insufficient funds
// 	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
// 	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
// 	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(sdk.NewCoin(NOLUS_DENOM, sdk.NewInt(10))))
// 	suite.Require().NoError(err)

// 	dfd := ante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, nil)
// 	dtd := keeper.NewDeductTaxDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.TaxKeeper)
// 	antehandler := sdk.ChainAnteDecorators(dfd, dtd)

// 	treasuryAddr, err := sdk.AccAddressFromBech32(suite.app.TaxKeeper.ContractAddress(suite.ctx))
// 	suite.Require().NoError(err)

// 	_, err = antehandler(suite.ctx, tx, false)
// 	suite.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

// 	// Set account with sufficient funds
// 	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
// 	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(sdk.NewCoin(NOLUS_DENOM, sdk.NewInt(200))))
// 	suite.Require().NoError(err)

// 	_, err = antehandler(suite.ctx, tx, false)
// 	suite.Require().Nil(err, "Tx errored after account has been set with sufficient funds")

// 	reqTax := sdk.NewCoins(sdk.NewCoin(NOLUS_DENOM, sdk.NewInt(60)))

// 	suite.EqualValues(reqTax, suite.app.BankKeeper.GetAllBalances(suite.ctx, treasuryAddr), "Tax differs from 40%")
// }

func (suite *KeeperTestSuite) TestTaxDecorator() {
	// todo check remaining fees are correctly distributed

	testCases := []struct {
		title              string
		feeDenoms          []string
		feeAmount          sdk.Int
		expPass            bool
		expTreasuryBalance sdk.Coins
		expErr             error
	}{
		{
			title:              "successful 40% tax application should increase the treasury balance",
			feeDenoms:          []string{NOLUS_DENOM},
			feeAmount:          sdk.NewInt(10),
			expPass:            true,
			expTreasuryBalance: sdk.NewCoins(sdk.NewCoin(NOLUS_DENOM, sdk.NewInt(4))),
			expErr:             nil,
		},
		{
			title:              "pay fees with insufficient funds should fail",
			feeDenoms:          []string{NOLUS_DENOM},
			feeAmount:          sdk.NewInt(100000),
			expPass:            false,
			expTreasuryBalance: nil,
			expErr:             sdkerrors.ErrInsufficientFunds,
		},
		// {
		// 	title:              "pay fees with less then minimum fee threshold should pass, but no tax collected",
		// 	feeDenoms:          []string{NOLUS_DENOM},
		// 	feeAmount:          sdk.NewInt(6),
		// 	expPass:            true,
		// 	expTreasuryBalance: sdk.NewCoins(sdk.NewCoin(NOLUS_DENOM, sdk.NewInt(0))),
		// 	expErr:             nil,
		// },
		{
			title:              "pay fees with not allowed denom should fail",
			feeDenoms:          []string{ATOM_DENOM},
			feeAmount:          sdk.NewInt(100),
			expPass:            false,
			expTreasuryBalance: nil,
			expErr:             types.ErrInvalidFeeDenom,
		},
		// {
		// 	title:              "pay fees with multiple denoms should fail",
		// 	feeDenoms:          []string{NOLUS_DENOM, ATOM_DENOM},
		// 	feeAmount:          sdk.NewInt(100),
		// 	expPass:            false,
		// 	expTreasuryBalance: nil,
		// 	expErr:             types.ErrTooManyFeeCoins,
		// },
		// {
		// 	title:              "tx without fees should fail",
		// 	feeDenoms:          []string{},
		// 	feeAmount:          sdk.NewInt(0),
		// 	expPass:            false,
		// 	expTreasuryBalance: nil,
		// 	expErr:             types.ErrFeesNotSet,
		// },
		// {
		// 	title:              "FALSE POSITIVE. tx without fees passes and do not increase the treasury",
		// 	feeDenoms:          []string{},
		// 	feeAmount:          sdk.NewInt(0),
		// 	expPass:            true,
		// 	expTreasuryBalance: sdk.NewCoins(sdk.NewCoin(NOLUS_DENOM, sdk.NewInt(0))),
		// 	// expErr: ?
		// },
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
			// startBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, NOLUS_DENOM)
			startBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)

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

				// endBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, NOLUS_DENOM)
				endBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
				suite.Require().Equal(startBalances, endBalances, "Start balances should be equal to end balances")
			}
		})
	}
}
