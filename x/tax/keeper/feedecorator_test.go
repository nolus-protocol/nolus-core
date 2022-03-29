package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/keeper"
)

func (suite *KeeperTestSuite) TestMempoolFeeDecoratorAnteHandle() {
	suite.SetupTest(true)
	baseDenom := sdk.DefaultBondDenom

	tests := []struct {
		name         string
		txFee        sdk.Coins
		minGasPrices sdk.DecCoins
		gasRequested uint64
		isCheckTx    bool
		expectPass   bool
	}{
		{
			name:         "no min gas price - checktx",
			txFee:        sdk.NewCoins(),
			minGasPrices: sdk.NewDecCoins(),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   true,
		},
		{
			name:         "no min gas price - delivertx",
			txFee:        sdk.NewCoins(),
			minGasPrices: sdk.NewDecCoins(),
			gasRequested: 10000,
			isCheckTx:    false,
			expectPass:   true,
		},
		{
			name:  "valid basedenom fee",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 1000,
			isCheckTx:    true,
			expectPass:   true,
		},
		{
			name:  "not enough fee in checktx",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   false,
		},
		{
			name:  "works with not enough fee in delivertx",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 10000,
			isCheckTx:    false,
			expectPass:   true,
		},
	}

	for _, tc := range tests {

		suite.ctx = suite.ctx.WithIsCheckTx(tc.isCheckTx)
		suite.ctx = suite.ctx.WithMinGasPrices(tc.minGasPrices)

		suite.txBuilder.SetFeeAmount(tc.txFee)
		suite.txBuilder.SetGasLimit(tc.gasRequested)

		tx := suite.txBuilder.GetTx()

		mfd := keeper.NewMempoolFeeDecorator(suite.app.TaxKeeper)
		antehandler := sdk.ChainAnteDecorators(mfd)
		_, err := antehandler(suite.ctx, tx, false)
		if tc.expectPass {
			suite.Require().NoError(err, "test: %s", tc.name)
		} else {
			suite.Require().Error(err, "test: %s", tc.name)
		}
	}

}

func (suite *KeeperTestSuite) TestApplyFee() {
	suite.SetupTest(true)
	baseDenom := sdk.DefaultBondDenom
	defaultFeeRate := int64(suite.app.TaxKeeper.FeeRate(suite.ctx))

	type expected struct {
		proceeds  sdk.Coins
		remaining sdk.Coins
		err       error
	}

	var testCases = []struct {
		name     string
		feeRate  int64
		feeCoins sdk.Coins
		expect   expected
	}{
		{
			name:     "works with no fee rate",
			feeRate:  0,
			feeCoins: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 50)),
			expect: expected{
				proceeds:  sdk.NewCoins(),
				remaining: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 50)),
				err:       nil,
			},
		},
		{
			name:     "works with default fee rate and enought coins",
			feeRate:  defaultFeeRate,
			feeCoins: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 50)),
			expect: expected{
				proceeds:  sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 20)),
				remaining: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 30)),
				err:       nil,
			},
		},
		{
			name:     "works with gready fee rate",
			feeRate:  100,
			feeCoins: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 50)),
			expect: expected{
				proceeds:  sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 50)),
				remaining: nil,
				err:       nil,
			},
		},
		{
			name:     "works with default fee rate and no coins",
			feeRate:  defaultFeeRate,
			feeCoins: sdk.NewCoins(),
			expect: expected{
				proceeds:  sdk.NewCoins(),
				remaining: sdk.NewCoins(),
				err:       nil,
			},
		},
	}

	for _, tc := range testCases {
		feeRate := sdk.NewDec(tc.feeRate)
		testName := fmt.Sprintf("test: %s", tc.name)

		actualProceeds, deductedFees, err := keeper.ApplyFee(feeRate, tc.feeCoins)

		if tc.expect.err == nil {
			suite.Require().NoError(err, testName)
		} else {
			suite.Require().Error(err, testName)
		}

		suite.EqualValues(tc.expect.proceeds, actualProceeds, testName)
		suite.EqualValues(tc.expect.remaining, deductedFees, testName)

		if !tc.feeCoins.Empty() {
			suite.EqualValues(tc.feeCoins, actualProceeds.Add(deductedFees...), testName)
		}
	}
}
