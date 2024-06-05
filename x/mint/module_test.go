package mint_test

import (
	"testing"

	"github.com/Nolus-Protocol/nolus-core/testutil/simapp"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	// simapp.TestSetup initializes the app
	app, err := simapp.TestSetup(t)
	require.NoError(t, err)
	ctx := app.BaseApp.NewContext(false)

	acc := app.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
