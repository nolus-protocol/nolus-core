package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	v087 "github.com/Nolus-Protocol/nolus-core/app/upgrades/v087"
	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
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
