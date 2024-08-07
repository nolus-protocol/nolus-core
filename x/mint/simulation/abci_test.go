package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/testutil/simapp"
	"github.com/Nolus-Protocol/nolus-core/x/mint"

	"github.com/stretchr/testify/require"
)

func Test_BeginBlock(t *testing.T) {
	params.GetDefaultConfig()
	app, err := simapp.TestSetup(t)
	if err != nil {
		t.Errorf("Error while creating simapp: %v\"", err)
	}
	blockTime := time.Now()
	ctx := app.BaseApp.NewContext(false).WithBlockTime(blockTime)
	minterKeeper := app.MintKeeper

	err = mint.BeginBlocker(ctx, *minterKeeper)
	require.NoError(t, err)

	header := tmproto.Header{Height: app.LastBlockHeight() + 2}
	ctx2 := ctx.WithBlockHeader(header).WithBlockTime(blockTime.Add(time.Second * 40))

	err = mint.BeginBlocker(ctx2, *minterKeeper)
	require.NoError(t, err)

	minter := minterKeeper.GetMinter(ctx2)
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx2, types.FeeCollectorName)
	feesCollectedInt := app.BankKeeper.GetAllBalances(ctx2, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)
	fmt.Printf("norm %v, total %v \n", minter.NormTimePassed, minter.TotalMinted)
	fmt.Printf("balance %v \n", feesCollected)
	require.Equal(t, math.NewIntFromBigInt(minter.TotalMinted.BigInt()), feesCollectedInt.AmountOf(sdk.DefaultBondDenom))
}
