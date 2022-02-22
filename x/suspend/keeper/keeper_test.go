package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	nolusapp "gitlab-nomo.credissimo.net/nomo/cosmzone/app"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"

	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	txBuilder client.TxBuilder
	clientCtx client.Context
	app       *nolusapp.App
}

func (suite *KeeperTestSuite) SetupTest(isCheckTx bool) {
	tempDir := suite.T().TempDir()
	suite.app, suite.ctx = nolusapp.CreateTestApp(isCheckTx, tempDir)
	suite.ctx = suite.ctx.WithBlockHeight(1)

	// Set up TxConfig.
	encodingConfig := simapp.MakeTestEncodingConfig()
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	suite.txBuilder.SetMsgs([]sdk.Msg{}...)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetSuspendState() {
	suite.SetupTest(true)

	encodingConfig := simapp.MakeTestEncodingConfig()
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)

	// set initial state
	_, _, adminAddr := sdktestutil.KeyTestPubAddr()
	state := types.NewSuspendedState(adminAddr.String(), false, suite.ctx.BlockHeight())
	suite.app.SuspendKeeper.SetState(suite.ctx, state)

	initialstate := suite.app.SuspendKeeper.GetState(suite.ctx)
	suite.Require().False(initialstate.Suspended)

	// try suspend with no admin adress
	_, _, addr1 := sdktestutil.KeyTestPubAddr()
	err := suite.app.SuspendKeeper.SetSuspendState(suite.ctx, true, addr1.String(), suite.ctx.BlockHeight())
	suite.Require().Error(err, err.Error())
	afterstate := suite.app.SuspendKeeper.GetState(suite.ctx)
	suite.Require().False(afterstate.Suspended)

	// try suspend with admin adress
	err = suite.app.SuspendKeeper.SetSuspendState(suite.ctx, true, adminAddr.String(), suite.ctx.BlockHeight())
	suite.Require().NoError(err)
	afterstate = suite.app.SuspendKeeper.GetState(suite.ctx)
	suite.Require().True(afterstate.Suspended)

	// try unsuspend with no admin adress
	err = suite.app.SuspendKeeper.SetSuspendState(suite.ctx, false, addr1.String(), suite.ctx.BlockHeight())
	suite.Require().Error(err, err.Error())
	afterstate = suite.app.SuspendKeeper.GetState(suite.ctx)
	suite.Require().True(afterstate.Suspended)

	// try unsuspend with admin adress
	err = suite.app.SuspendKeeper.SetSuspendState(suite.ctx, false, adminAddr.String(), suite.ctx.BlockHeight())
	suite.Require().NoError(err)
	afterstate = suite.app.SuspendKeeper.GetState(suite.ctx)
	suite.Require().False(afterstate.Suspended)

}
