package mint_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	authtypes "cosmossdk.io/x/auth/types"
	"github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	// simapp.Setup initializes the app
	nolusApp := New(t, app.DefaultNodeHome, true)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	acc := app.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
