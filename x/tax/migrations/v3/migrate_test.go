package v3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/x/tax"
	"github.com/Nolus-Protocol/nolus-core/x/tax/exported"
	v3 "github.com/Nolus-Protocol/nolus-core/x/tax/migrations/v3"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
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
	encCfg := moduletestutil.MakeTestEncodingConfig(tax.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(v3.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	legacySubspace := newMockSubspace(types.DefaultParams())
	storeService := runtime.NewKVStoreService(storeKey)
	store := storeService.OpenKVStore(ctx)

	require.NoError(t, v3.Migrate(ctx, store, legacySubspace, cdc))

	b, err := store.Get(v3.ParamsKey)
	require.NoError(t, err)

	var res types.Params
	require.NoError(t, cdc.Unmarshal(b, &res))
	require.Equal(t, legacySubspace.ps, res)
}
