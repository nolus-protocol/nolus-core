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

type KeeperTestCase struct {
	name        string
	expectPass  bool
	fromAddress string
	suspended   bool
}

func (suite *KeeperTestSuite) setInitialState(suspended bool) sdk.AccAddress {
	// set initial state
	_, _, adminAddr := sdktestutil.KeyTestPubAddr()
	state := types.NewSuspendedState(adminAddr.String(), suspended, suite.ctx.BlockHeight())
	suite.app.SuspendKeeper.SetState(suite.ctx, state)

	initialstate := suite.app.SuspendKeeper.GetState(suite.ctx)
	suite.Require().Equal(suspended, initialstate.Suspended)

	return adminAddr
}

func (suite *KeeperTestSuite) TestSetSuspendState() {
	suite.SetupTest(true)

	// no admin address is set
	_, _, addr1 := sdktestutil.KeyTestPubAddr()
	err := suite.app.SuspendKeeper.SetSuspendState(suite.ctx, true, addr1.String(), suite.ctx.BlockHeight())
	suite.Require().EqualError(err, "No admin address is set: unauthorized")

	// set initial state
	adminAddr := suite.setInitialState(false)

	tests := []KeeperTestCase{
		{
			name:        "suspend with empty adress",
			expectPass:  false,
			fromAddress: "",
			suspended:   true,
		},
		{
			name:        "suspend with invalid adress",
			expectPass:  false,
			fromAddress: "invalidaddres",
			suspended:   true,
		},
		{
			name:        "suspend with no admin adress",
			expectPass:  false,
			fromAddress: addr1.String(),
			suspended:   true,
		},
		{
			name:        "suspend with admin adress",
			expectPass:  true,
			fromAddress: adminAddr.String(),
			suspended:   true,
		},
		{
			name:        "unsuspend with no admin adress",
			expectPass:  false,
			fromAddress: addr1.String(),
			suspended:   false,
		},
		{
			name:        "unsuspend with admin adress",
			expectPass:  true,
			fromAddress: adminAddr.String(),
			suspended:   false,
		},
	}
	for _, tc := range tests {
		err = suite.app.SuspendKeeper.SetSuspendState(suite.ctx, tc.suspended, tc.fromAddress, suite.ctx.BlockHeight())
		if tc.expectPass {
			suite.Require().NoError(err, "test: %s", tc.name)
		} else {
			suite.Require().Error(err, "test: %s ; error: %s", tc.name, err.Error())
		}

		afterstate := suite.app.SuspendKeeper.GetState(suite.ctx)
		if tc.expectPass {
			suite.Require().Equal(tc.suspended, afterstate.Suspended)
		} else {
			suite.Require().NotEqual(tc.suspended, afterstate.Suspended)
		}
	}

}
