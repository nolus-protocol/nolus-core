package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	simulationapp "github.com/Nolus-Protocol/nolus-core/testutil/simapp"
	"github.com/Nolus-Protocol/nolus-core/x/mint/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

var (
	defaultMintDenom              = sdk.DefaultBondDenom
	defaultMaxMintableNanoseconds = uint64(60000000000)
)

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *KeeperTestSuite) SetupTest(isCheckTx bool) {
	var err error
	_ = params.GetDefaultConfig()
	s.app, err = simulationapp.TestSetup(s.T())
	s.Require().NoError(err)

	blockTime := time.Now()
	header := tmproto.Header{Height: s.app.LastBlockHeight() + 1}
	s.ctx = s.app.BaseApp.NewContext(false, header).WithBlockTime(blockTime)
	s.msgServer = keeper.NewMsgServerImpl(*s.app.MintKeeper)
}

func (s *KeeperTestSuite) TestParamsQuery() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	resp, err := minterKeeper.Params(s.ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(defaultMintDenom, resp.Params.MintDenom)
	s.Require().Equal(sdkmath.NewUint(defaultMaxMintableNanoseconds), resp.Params.MaxMintableNanoseconds)
}

func (s *KeeperTestSuite) TestMintState() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	resp, err := minterKeeper.MintState(s.ctx, &types.QueryMintStateRequest{})
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.ZeroUint(), resp.TotalMinted)
}
