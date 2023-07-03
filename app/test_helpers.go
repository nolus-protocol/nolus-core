package app

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxtypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	dbm "github.com/tendermint/tm-db"
)

// returns context and app with params set on account keeper.
func CreateTestApp(isCheckTx bool, tempDir string) (*App, sdk.Context) {
	encoding := MakeEncodingConfig(ModuleBasics)

	app := New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, map[int64]bool{},
		tempDir, simapp.FlagPeriodValue, encoding,
		simapp.EmptyAppOptions{})

	// cosmoscmd.SetPrefixes(nolusapp.AccountAddressPrefix)
	// sdk.GetConfig().SetBech32PrefixForAccount(nolusapp.AccountAddressPrefix, nolusapp.AccountAddressPrefixPub)
	params.SetAddressPrefixes()

	testapp := app

	ctx := testapp.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	testapp.TaxKeeper.SetParams(ctx, taxtypes.DefaultParams())
	testapp.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	testapp.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	testapp.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	return testapp, ctx
}
