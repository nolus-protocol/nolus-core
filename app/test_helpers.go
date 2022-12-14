package app

import (
	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/spm/cosmoscmd"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/tendermint/tendermint/libs/log"

	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxtypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"
	dbm "github.com/tendermint/tm-db"
)

// returns context and app with params set on account keeper.
func CreateTestApp(isCheckTx bool, tempDir string) (*App, sdk.Context) {
	encoding := cosmoscmd.MakeEncodingConfig(ModuleBasics)

	app := New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, map[int64]bool{},
		tempDir, simapp.FlagPeriodValue, encoding,
		simapp.EmptyAppOptions{})

	// cosmoscmd.SetPrefixes(nolusapp.AccountAddressPrefix)
	// sdk.GetConfig().SetBech32PrefixForAccount(nolusapp.AccountAddressPrefix, nolusapp.AccountAddressPrefixPub)
	params.SetAddressPrefixes()

	testapp := app.(*App)

	ctx := testapp.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	testapp.TaxKeeper.SetParams(ctx, taxtypes.DefaultParams())
	testapp.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	testapp.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	testapp.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	testapp.FeeKeeper.SetParams(ctx, feetypes.DefaultParams())

	return testapp, ctx
}
