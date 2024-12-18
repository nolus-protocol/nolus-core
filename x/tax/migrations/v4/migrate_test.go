package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/x/tax"
	v4 "github.com/Nolus-Protocol/nolus-core/x/tax/migrations/v4"
	typesv1 "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func TestMigrate(t *testing.T) {
	params.GetDefaultConfig()
	encCfg := moduletestutil.MakeTestEncodingConfig(tax.AppModuleBasic{})
	cdc := encCfg.Codec

	oldParams := typesv1.Params{
		FeeRate:         types.DefaultFeeRate,
		ContractAddress: types.DefaultTreasuryAddress,
		BaseDenom:       types.DefaultBaseDenom,
		FeeParams:       []*typesv1.FeeParam{},
	}

	expectedParams := types.Params{
		FeeRate:         types.DefaultFeeRate,
		TreasuryAddress: types.DefaultTreasuryAddress,
		BaseDenom:       types.DefaultBaseDenom,
		DexFeeParams:    v4.DexFeeParams,
	}

	storeKey := storetypes.NewKVStoreKey(v4.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	storeService := runtime.NewKVStoreService(storeKey)
	store := storeService.OpenKVStore(ctx)
	store.Set(types.ParamsKey, cdc.MustMarshal(&oldParams))

	require.NoError(t, v4.Migrate(ctx, store, cdc))

	b, err := store.Get(types.ParamsKey)
	require.NoError(t, err)

	var res types.Params
	require.NoError(t, cdc.Unmarshal(b, &res))
	require.Equal(t, expectedParams, res)
}
