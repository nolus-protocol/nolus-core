package app

import (
	"cosmossdk.io/log"
	db "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxtypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
)

// returns context and app with params set on account keeper.
func CreateTestApp(isCheckTx bool, tempDir string) (*App, sdk.Context) {
	app := New(
		log.NewNopLogger(),
		db.NewMemDB(),
		nil,
		true,
		sims.EmptyAppOptions{},
	)

	params.GetDefaultConfig()

	testapp := app
	ctx := testapp.BaseApp.NewContext(isCheckTx)
	_ = testapp.TaxKeeper.SetParams(ctx, taxtypes.DefaultParams())
	_ = testapp.MintKeeper.SetParams(ctx, minttypes.DefaultParams())

	err := testapp.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams())
	if err != nil {
		panic(err)
	}

	err = testapp.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	if err != nil {
		panic(err)
	}
	return testapp, ctx
}
