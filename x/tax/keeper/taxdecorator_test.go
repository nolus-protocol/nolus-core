package keeper_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/keeper"
)

func (suite *KeeperTestSuite) TestTaxes() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := sdktestutil.KeyTestPubAddr()

	// msg and signatures
	msg := sdktestutil.NewTestMsg(addr1)
	feeAmount := sdktestutil.NewTestFeeAmount() // 150atom
	gasLimit := sdktestutil.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	// Set account with insufficient funds
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10))))
	suite.Require().NoError(err)

	dfd := ante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, nil)
	dtd := keeper.NewDeductTaxDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.TaxKeeper)
	antehandler := sdk.ChainAnteDecorators(dfd, dtd)

	treasuryAddr, err := sdk.AccAddressFromBech32(suite.app.TaxKeeper.ContractAddress(suite.ctx))
	suite.Require().NoError(err)

	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200))))
	suite.Require().NoError(err)

	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Nil(err, "Tx errored after account has been set with sufficient funds")

	reqTax := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(60)))

	suite.EqualValues(reqTax, suite.app.BankKeeper.GetAllBalances(suite.ctx, treasuryAddr), "Tax differs from 40%")
}
