package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/keeper"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

func (suite *KeeperTestSuite) TestSuspendAnteHandle() {
	suite.SetupTest(true)

	encodingConfig := simapp.MakeTestEncodingConfig()
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)

	// set initial state to suspended
	_, _, adminAddr := sdktestutil.KeyTestPubAddr()
	state := types.NewSuspendedState(adminAddr.String(), true, suite.ctx.BlockHeight())
	suite.app.SuspendKeeper.SetState(suite.ctx, state)

	initialstate := suite.app.SuspendKeeper.GetState(suite.ctx)
	suite.Require().True(initialstate.Suspended)

	sd := keeper.NewSuspendDecorator(suite.app.SuspendKeeper)
	antehandler := sdk.ChainAnteDecorators(sd)

	newCtx := suite.ctx.WithBlockHeight(10)

	// send random message => should fail
	suite.txBuilder.SetMsgs([]sdk.Msg{}...)
	tx := suite.txBuilder.GetTx()
	_, err := antehandler(newCtx, tx, false)
	suite.Require().Error(err, err.Error())

	// send unsuspend message => should pass
	unsuspendMsg := types.NewMsgUnsuspend(adminAddr.String())
	suite.txBuilder.SetMsgs(unsuspendMsg)
	tx = suite.txBuilder.GetTx()
	_, err = antehandler(newCtx, tx, false)
	suite.Require().NoError(err)

}
