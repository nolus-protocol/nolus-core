package mint_test

// refactor: fix when simulation refactoring is done
// import (
// 	"testing"

// 	abcitypes "github.com/cometbft/cometbft/abci/types"
// 	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
// 	"github.com/cosmos/cosmos-sdk/testutil/sims"
// 	"github.com/stretchr/testify/require"

// 	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
// 	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
// )

// func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
// 	app := sims.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

// 	app.InitChain(
// 		abcitypes.RequestInitChain{
// 			AppStateBytes: []byte("{}"),
// 			ChainId:       "test-chain-id",
// 		},
// 	)

// 	acc := app.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
// 	require.NotNil(t, acc)
// }
