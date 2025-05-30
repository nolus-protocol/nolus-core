package v2_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Nolus-Protocol/nolus-core/testutil"
	v2 "github.com/Nolus-Protocol/nolus-core/x/contractmanager/migrations/v2"
	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
	typesv1 "github.com/Nolus-Protocol/nolus-core/x/contractmanager/types/v1"
)

type V2ContractManagerMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V2ContractManagerMigrationTestSuite))
}

func (suite *V2ContractManagerMigrationTestSuite) TestFailuresUpgrade() {
	var (
		app      = suite.GetNolusZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	addressOne := testutil.TestOwnerAddress
	addressTwo := "nolus17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgsmc5xhq"

	// Write old state
	storeService := runtime.NewKVStoreService(storeKey)
	store := storeService.OpenKVStore(ctx)
	var i uint64
	for i = 0; i < 4; i++ {
		var addr string
		if i < 2 {
			addr = addressOne
		} else {
			addr = addressTwo
		}
		failure := typesv1.Failure{
			ChannelId: "channel-0",
			Address:   addr,
			Id:        i % 2,
			AckType:   types.Ack,
		}
		bz := cdc.MustMarshal(&failure)
		store.Set(types.GetFailureKey(failure.Address, failure.Id), bz)
	}

	// Run migration
	suite.NoError(v2.MigrateStore(ctx, storeService))

	// Check elements should be empty
	expected := app.ContractManagerKeeper.GetAllFailures(ctx)
	suite.Require().ElementsMatch(expected, []types.Failure{})

	// Non-existent returns error
	_, err := app.ContractManagerKeeper.GetFailure(ctx, sdk.MustAccAddressFromBech32(addressTwo), 0)
	suite.Require().Error(err)

	// Check next id key is reset
	oneKey := app.ContractManagerKeeper.GetNextFailureIDKey(ctx, addressOne)
	suite.Require().Equal(oneKey, uint64(0))
	twoKey := app.ContractManagerKeeper.GetNextFailureIDKey(ctx, addressTwo)
	suite.Require().Equal(twoKey, uint64(0))
}
