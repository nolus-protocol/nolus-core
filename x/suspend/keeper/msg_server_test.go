package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/keeper"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

func (suite *KeeperTestSuite) TestMsgServer() {
	suite.SetupTest(true)

	// set initial state
	adminAddr := suite.setInitialState(false)

	msgServer := keeper.NewMsgServerImpl(suite.app.SuspendKeeper)
	goCtx := sdk.WrapSDKContext(suite.ctx)

	resp1, err := msgServer.Suspend(goCtx, types.NewMsgSuspend(adminAddr.String(), true, suite.ctx.BlockHeight()))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp1)

	resp2, err := msgServer.Unsuspend(goCtx, types.NewMsgUnsuspend(adminAddr.String()))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp2)

}
