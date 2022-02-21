package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/tendermint/tendermint/libs/log"

	nolusapp "gitlab-nomo.credissimo.net/nomo/cosmzone/app"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/spm/cosmoscmd"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	minttypes "gitlab-nomo.credissimo.net/nomo/cosmzone/x/mint/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/keeper"
	taxtypes "gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/types"
)

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	acc  authtypes.AccountI
	priv cryptotypes.PrivKey
}

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	suite.Suite

	app         *nolusapp.App
	ctx         sdk.Context
	clientCtx   client.Context
	txBuilder   client.TxBuilder
	anteHandler sdk.AnteHandler
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool, tempDir string) (*nolusapp.App, sdk.Context) {
	encoding := cosmoscmd.MakeEncodingConfig(nolusapp.ModuleBasics)

	app := nolusapp.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, map[int64]bool{},
		tempDir, simapp.FlagPeriodValue, encoding,
		simapp.EmptyAppOptions{})

	// cosmoscmd.SetPrefixes(nolusapp.AccountAddressPrefix)
	// sdk.GetConfig().SetBech32PrefixForAccount(nolusapp.AccountAddressPrefix, nolusapp.AccountAddressPrefixPub)
	params.SetAddressPrefixes()

	testapp := app.(*nolusapp.App)

	ctx := testapp.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	testapp.TaxKeeper.SetParams(ctx, taxtypes.DefaultParams())
	testapp.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	testapp.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	testapp.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	return testapp, ctx
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (suite *AnteTestSuite) SetupTest(isCheckTx bool) {
	tempDir := suite.T().TempDir()
	suite.app, suite.ctx = createTestApp(isCheckTx, tempDir)
	suite.ctx = suite.ctx.WithBlockHeight(1)

	// Set up TxConfig.
	encodingConfig := simapp.MakeTestEncodingConfig()
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	suite.txBuilder.SetMsgs([]sdk.Msg{}...)

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   suite.app.AccountKeeper,
			BankKeeper:      suite.app.BankKeeper,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)

	suite.Require().NoError(err)
	suite.anteHandler = anteHandler
}

// CreateTestAccounts creates `numAccs` accounts, and return all relevant
// information about them including their private keys.
func (suite *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := sdktestutil.KeyTestPubAddr()
		println("addr: ", addr.String())
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
		err := acc.SetAccountNumber(uint64(i))
		suite.Require().NoError(err)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		someCoins := sdk.Coins{
			sdk.NewInt64Coin("nolus", 10000000),
		}

		fmt.Printf("Mint %d nolus from module %s \n", someCoins.AmountOf("nolus"), minttypes.ModuleName)

		err = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, someCoins)
		suite.Require().NoError(err)

		modulacc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, minttypes.ModuleName)
		moduleAddrr := modulacc.GetAddress()
		println("Module address: ", moduleAddrr.String())

		moduleBalance := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddrr, "nolus")
		println("Balance module: ", strconv.Itoa(int(moduleBalance.Amount.Int64())))

		err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, someCoins)
		suite.Require().NoError(err)

		addrBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nolus")
		println("Balanace: ", strconv.Itoa(int(addrBalance.Amount.Int64())))

		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			suite.txBuilder, priv, suite.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return suite.txBuilder.GetTx(), nil
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
