package keeper_test

import (
	"github.com/Nolus-Protocol/nolus-core/x/mint/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (s *KeeperTestSuite) TestQueryParams() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	querierFunc := keeper.NewQuerier(minterKeeper, codec.NewLegacyAmino())
	bytes, err := querierFunc(s.ctx, []string{types.QueryParameters}, abci.RequestQuery{})

	s.Require().NoError(err)
	s.Require().Equal("{\n  \"mint_denom\": \"stake\",\n  \"max_mintable_nanoseconds\": \"60000000000\"\n}", string(bytes))
}

func (s *KeeperTestSuite) TestQueryMintState() {
	s.SetupTest(false)
	minterKeeper := s.app.MintKeeper

	querierFunc := keeper.NewQuerier(minterKeeper, codec.NewLegacyAmino())
	bytes, err := querierFunc(s.ctx, []string{types.QueryMintState}, abci.RequestQuery{})

	s.Require().NoError(err)
	s.Require().Equal("{\n  \"norm_time_passed\": \"0.470000000000000000\",\n  \"total_minted\": \"0\"\n}", string(bytes))
}
