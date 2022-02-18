package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/tendermint/tendermint/libs/log"

	nolusapp "gitlab-nomo.credissimo.net/nomo/cosmzone/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/spm/cosmoscmd"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/keeper"
	taxtypes "gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/types"
)

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	suite.Suite

	app       *nolusapp.App
	ctx       sdk.Context
	clientCtx client.Context
	txBuilder client.TxBuilder
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool, tempDir string) (*nolusapp.App, sdk.Context) {
	encoding := cosmoscmd.MakeEncodingConfig(nolusapp.ModuleBasics)

	app := nolusapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, map[int64]bool{},
		tempDir, simapp.FlagPeriodValue, encoding,
		simapp.EmptyAppOptions{})

	testapp := app.(*nolusapp.App)
	ctx := testapp.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	testapp.TaxKeeper.SetParams(ctx, taxtypes.DefaultParams())

	return testapp, ctx
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (suite *AnteTestSuite) SetupTest(isCheckTx bool) {
	tempDir := suite.T().TempDir()
	suite.app, suite.ctx = createTestApp(isCheckTx, tempDir)
	suite.ctx = suite.ctx.WithBlockHeight(1)

	// Set up TxConfig.
	encodingConfig := simapp.MakeTestEncodingConfig()

	suite.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	suite.txBuilder.SetMsgs([]sdk.Msg{}...)
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func (suite *AnteTestSuite) TestApplyFee() {
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
