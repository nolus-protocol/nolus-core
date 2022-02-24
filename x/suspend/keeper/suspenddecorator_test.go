package keeper_test

import (
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/keeper"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

type DecoratorTestCase struct {
	name            string
	messages        []sdk.Msg
	newCtx          sdk.Context
	expectPass      bool
	expectSuspended bool
}

func (suite *KeeperTestSuite) TestSuspendAnteHandle() {
	suite.SetupTest(true)

	// set initial state to suspended
	adminAddr := suite.setInitialState(true)
	_, _, addr1 := sdktestutil.KeyTestPubAddr()
	sd := keeper.NewSuspendDecorator(suite.app.SuspendKeeper)
	antehandler := sdk.ChainAnteDecorators(sd)

	tests := []DecoratorTestCase{
		{
			name:            "send empty message, no admin => should fail",
			messages:        []sdk.Msg{},
			newCtx:          suite.ctx.WithBlockHeight(10),
			expectPass:      false,
			expectSuspended: true,
		},
		{
			name:            "send unsuspend message, admin => should pass",
			messages:        []sdk.Msg{types.NewMsgUnsuspend(adminAddr.String())},
			newCtx:          suite.ctx.WithBlockHeight(10),
			expectPass:      false,
			expectSuspended: true,
		},
		{
			name: "send multiple messages, including unsuspend",
			messages: []sdk.Msg{
				sdktestutil.NewTestMsg(adminAddr),
				types.NewMsgUnsuspend(adminAddr.String()),
				sdktestutil.NewTestMsg(addr1),
			},
			newCtx:          suite.ctx.WithBlockHeight(10),
			expectPass:      true,
			expectSuspended: false,
		},
	}
	for _, tc := range tests {
		suite.txBuilder.SetMsgs(tc.messages...)
		tx := suite.txBuilder.GetTx()
		_, err := antehandler(tc.newCtx, tx, false)

		if tc.expectPass {
			suite.Require().NoError(err, "test: %s", tc.name)
		} else {
			suite.Require().Error(err, "test: %s ; error: %s", tc.name, err.Error())
		}

		afterstate := suite.app.SuspendKeeper.GetState(suite.ctx)
		suite.Require().Equal(tc.expectSuspended, afterstate.Suspended)
	}

}
