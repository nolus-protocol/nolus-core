package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/suite"

	nolusapp "github.com/Nolus-Protocol/nolus-core/app"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	app       *nolusapp.App
	clientCtx client.Context
	txBuilder client.TxBuilder
}

func (s *KeeperTestSuite) SetupTest(isCheckTx bool) {
	tempDir := s.T().TempDir()
	s.app, s.ctx = nolusapp.CreateTestApp(isCheckTx, tempDir)
	s.ctx = s.ctx.WithBlockHeight(1)

	// set up TxConfig
	encodingConfig := moduletestutil.MakeTestEncodingConfig()
	s.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
	s.Require().NoError(s.txBuilder.SetMsgs([]sdk.Msg{}...))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
