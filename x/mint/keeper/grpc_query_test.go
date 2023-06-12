package keeper_test

import (
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	simulationapp "github.com/Nolus-Protocol/nolus-core/testutil/simapp"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

var (
	defaultMintDenom              = sdk.DefaultBondDenom
	defaultMaxMintableNanoseconds = uint64(60000000000)
)

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *KeeperTestSuite) SetupTest(isCheckTx bool) {
	var err error
	params.SetAddressPrefixes()
	s.app, err = simulationapp.TestSetup()
	s.Require().NoError(err)

	blockTime := time.Now()
	header := tmproto.Header{Height: s.app.LastBlockHeight() + 1}
	s.ctx = s.app.BaseApp.NewContext(false, header).WithBlockTime(blockTime)
	s.sdkWrappedCtx = sdk.WrapSDKContext(s.ctx)
}

func (s *KeeperTestSuite) TestParams() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	resp, err := minterKeeper.Params(s.sdkWrappedCtx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(defaultMintDenom, resp.Params.MintDenom)
	s.Require().Equal(sdk.NewUint(defaultMaxMintableNanoseconds), resp.Params.MaxMintableNanoseconds)
}

func (s *KeeperTestSuite) TestMintState() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	resp, err := minterKeeper.MintState(s.sdkWrappedCtx, &types.QueryMintStateRequest{})
	s.Require().NoError(err)
	s.Require().Equal(sdk.ZeroUint(), resp.TotalMinted)
}
