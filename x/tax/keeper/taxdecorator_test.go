package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/Nolus-Protocol/nolus-core/x/tax/keeper"
)

func (suite *KeeperTestSuite) TestTaxDecorator() {
	suite.SetupTest(true)

	HUNDRED_DEC := sdkmath.LegacyNewDec(100)
	const rnDenom = "atom"
	baseDenom := suite.app.TaxKeeper.GetParams(suite.ctx).BaseDenom

	testCases := []struct {
		title           string
		feeDenoms       []string
		feeAmount       sdkmath.Int
		feeRate         int32
		baseDenom       string
		contractAddress string
		expPass         bool
		expErr          error
	}{
		{
			title:           "successful tax deduction should increase the treasury balance",
			feeDenoms:       []string{baseDenom},
			feeAmount:       sdkmath.NewInt(100),
			feeRate:         40,
			baseDenom:       baseDenom,
			contractAddress: "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz",
			expPass:         true,
			expErr:          nil,
		},
		// {
		// 	title:     "tx with 0 fee rate should not increase the treasury balance",
		// 	feeDenoms: []string{baseDenom},
		// 	feeAmount: sdk.NewInt(100),
		// 	feeRate:   0,
		// 	expPass:   true,
		// 	expErr:    nil,
		// },
		// {
		// 	title:     "tx with tax is less then 1 should not increase the treasury balance",
		// 	feeDenoms: []string{baseDenom},
		// 	feeAmount: sdk.NewInt(1),
		// 	feeRate:   40,
		// 	expPass:   true,
		// 	expErr:    nil,
		// },
		// {
		// 	title:     "tx without fees should continue to the next AnteHandler",
		// 	feeDenoms: []string{},
		// 	feeAmount: sdk.NewInt(0),
		// 	feeRate:   40,
		// 	expPass:   true,
		// 	expErr:    nil,
		// },
		// {
		// 	title:     "pay fees with insufficient funds should fail",
		// 	feeDenoms: []string{baseDenom},
		// 	feeAmount: sdk.NewInt(100000),
		// 	feeRate:   40,
		// 	expPass:   false,
		// 	expErr:    sdkerrors.ErrInsufficientFunds,
		// },
		// {
		// 	title:     "pay fees with not allowed denom should fail",
		// 	feeDenoms: []string{rnDenom},
		// 	feeAmount: sdk.NewInt(100),
		// 	feeRate:   40,
		// 	expPass:   false,
		// 	expErr:    types.ErrInvalidFeeDenom,
		// },
		// {
		// 	title:     "pay fees with multiple denoms should fail",
		// 	feeDenoms: []string{baseDenom, rnDenom},
		// 	feeAmount: sdk.NewInt(100),
		// 	feeRate:   40,
		// 	expPass:   false,
		// 	expErr:    types.ErrTooManyFeeCoins,
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

			// set fee rate
			params := suite.app.TaxKeeper.GetParams(suite.ctx)
			params.FeeRate = tc.feeRate
			params.BaseDenom = tc.baseDenom
			params.ContractAddress = tc.contractAddress
			suite.app.TaxKeeper.SetParams(suite.ctx, params)

			// get chained ante handler
			dfd := ante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, nil, nil)
			dtd := keeper.NewDeductTaxDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, *suite.app.TaxKeeper)
			anteHandler := sdk.ChainAnteDecorators(dfd, dtd)

			// retrieve treasury address
			treasuryAddr, err := sdk.AccAddressFromBech32(params.ContractAddress)
			suite.Require().NoError(err)

			// add coins to pay the tax
			var txFees sdk.Coins
			for _, feeDenom := range tc.feeDenoms {
				txFees = txFees.Add(sdk.NewCoin(feeDenom, tc.feeAmount))
			}

			suite.txBuilder.SetFeeAmount(txFees)

			// call the ante handler
			_, err = anteHandler(suite.ctx, tx, false)
			if !tc.expPass {
				suite.Require().Error(err, "test: %s", tc.title)
				suite.ErrorIs(err, tc.expErr, tc.title)
				return
			}

			// pass is expected
			suite.Require().NoError(err, "test: %s", tc.title)

			expTreasuryBalance := sdk.Coins{} // empty treasury
			treasuryBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, treasuryAddr)
			feeRate := sdkmath.LegacyNewDec(int64(tc.feeRate))
			tax := feeRate.MulInt(tc.feeAmount).Quo(HUNDRED_DEC).TruncateInt()

			if txFees.Empty() || tc.feeRate == 0 || tax.LT(sdkmath.NewInt(1)) {
				suite.Require().Equal(expTreasuryBalance, treasuryBalance, "Treasury should be empty")
				return
			}

			feeDenom := tc.feeDenoms[0]
			expTreasuryBalance = expTreasuryBalance.Add(
				sdk.NewCoin(feeDenom, tax),
			)

			suite.Require().Equal(expTreasuryBalance, treasuryBalance, "Treasury should have collected correct tax amount")
		})
	}
}
