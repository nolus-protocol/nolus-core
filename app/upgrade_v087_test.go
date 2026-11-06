package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	v087 "github.com/Nolus-Protocol/nolus-core/app/upgrades/v087"
	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
	solanacarriermodule "github.com/Nolus-Protocol/nolus-core/x/solanacarrier"
	solanacarrierv2 "github.com/Nolus-Protocol/nolus-core/x/solanacarrier/migrations/v2"
)

func TestV087IsRegisteredWithoutStoreChanges(t *testing.T) {
	require.Equal(t, "v0.8.7", v087.UpgradeName)
	require.Empty(t, v087.Upgrade.StoreUpgrades.Added)
	require.Empty(t, v087.Upgrade.StoreUpgrades.Deleted)
	require.Empty(t, v087.Upgrade.StoreUpgrades.Renamed)

	names := make([]string, 0, len(Upgrades))
	for _, u := range Upgrades {
		names = append(names, u.UpgradeName)
	}
	require.Contains(t, names, v087.UpgradeName, "v0.8.7 must be registered in the Upgrades list")

	testApp, _ := CreateTestApp(true, t.TempDir())
	require.True(t, testApp.UpgradeKeeper.HasHandler(v087.UpgradeName),
		"the upgrade keeper must hold a handler for v0.8.7")
}

// Dropping one module from the input map proves the handler really runs the
// module migrations rather than echoing the map it was given.
func TestV087UpgradeHandlerReturnsTheMigratedVersionMap(t *testing.T) {
	testApp, ctx := CreateTestApp(true, t.TempDir())

	fromVM := testApp.mm.GetVersionMap()
	delete(fromVM, solanacarrier.ModuleName)

	handler := v087.CreateUpgradeHandler(
		testApp.mm,
		testApp.configurator,
		&testApp.AppKeepers,
		testApp.appCodec,
	)

	vm, err := handler(ctx, upgradetypes.Plan{Name: v087.UpgradeName}, fromVM)
	require.NoError(t, err)
	require.Equal(t, testApp.mm.GetVersionMap(), vm)
}

func TestV087UpgradeHandlerMigratesSolanaCarrierParamsToV2(t *testing.T) {
	testApp, ctx := CreateTestApp(true, t.TempDir())

	// Off curve and not the default fee payer, so keeping it is told apart from a reset.
	v1StoredFeePayer := []byte{
		0x4a, 0x7c, 0x9e, 0x0b, 0x4d, 0xdd, 0x06, 0x9c,
		0x64, 0x31, 0x08, 0x5d, 0x1c, 0x0b, 0x1e, 0x2f,
		0x2b, 0x5b, 0x3d, 0x9a, 0x84, 0x1f, 0x5e, 0x11,
		0xa2, 0xc7, 0x6d, 0x38, 0x90, 0x44, 0x1b, 0x00,
	}
	v1Params := append([]byte{0x0a, byte(len(v1StoredFeePayer))}, v1StoredFeePayer...) // field 1, fee_payer, alone
	ctx.KVStore(testApp.GetKey(solanacarrier.StoreKey)).Set(solanacarrier.ParamsKey, v1Params)

	fromVM := testApp.mm.GetVersionMap()
	fromVM[solanacarrier.ModuleName] = 1

	handler := v087.CreateUpgradeHandler(
		testApp.mm,
		testApp.configurator,
		&testApp.AppKeepers,
		testApp.appCodec,
	)

	vm, err := handler(ctx, upgradetypes.Plan{Name: v087.UpgradeName}, fromVM)
	require.NoError(t, err)
	require.Equal(t, uint64(2), vm[solanacarrier.ModuleName], "the upgrade must record solanacarrier at version 2")
	require.Equal(t, uint64(solanacarriermodule.ConsensusVersion), vm[solanacarrier.ModuleName])

	params, err := testApp.SolanaCarrierKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, solanacarrier.Params{
		FeePayer:        v1StoredFeePayer,
		AllowedPrograms: solanacarrierv2.AllowedPrograms(),
	}, params)
}
