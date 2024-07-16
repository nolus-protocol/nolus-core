package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/x/mint"
	"github.com/Nolus-Protocol/nolus-core/x/mint/exported"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	v2 "github.com/Nolus-Protocol/nolus-core/x/mint/migrations/v2"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps exported.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	params.GetDefaultConfig()
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(v2.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	legacySubspace := newMockSubspace(types.DefaultParams())
	storeService := runtime.NewKVStoreService(storeKey)
	store := storeService.OpenKVStore(ctx)

	require.NoError(t, v2.Migrate(ctx, runtime.NewKVStoreService(storeKey), legacySubspace, cdc))

	b, err := store.Get(v2.ParamsKey)
	require.NoError(t, err)

	var res types.Params
	require.NoError(t, cdc.Unmarshal(b, &res))
	require.Equal(t, legacySubspace.ps, res)
}
