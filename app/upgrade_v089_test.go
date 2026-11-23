package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	v089 "github.com/Nolus-Protocol/nolus-core/app/upgrades/v089"
	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

func TestV089IsRegisteredWithoutStoreChanges(t *testing.T) {
	require.Equal(t, "v0.8.9", v089.UpgradeName, "the on-chain plan name must be v0.8.9")
	require.Empty(t, v089.Upgrade.StoreUpgrades.Added, "v0.8.9 adds no stores")
	require.Empty(t, v089.Upgrade.StoreUpgrades.Deleted, "v0.8.9 deletes no stores")
	require.Empty(t, v089.Upgrade.StoreUpgrades.Renamed, "v0.8.9 renames no stores")

	names := make([]string, 0, len(Upgrades))
	for _, u := range Upgrades {
		names = append(names, u.UpgradeName)
	}
	require.Contains(t, names, v089.UpgradeName, "v0.8.9 must be registered in the Upgrades list")

	testApp, _ := CreateTestApp(true, t.TempDir())
	require.True(t, testApp.UpgradeKeeper.HasHandler(v089.UpgradeName),
		"the upgrade keeper must hold a handler for v0.8.9")
}

func TestV089UpgradeHandlerReturnsCurrentVersions(t *testing.T) {
	tests := []struct {
		name    string
		missing []string
	}{
		{name: "every module already at its current version"},
		// the handler must restore it, so it cannot just echo its input
		{name: "a module missing from the version map", missing: []string{solanacarrier.ModuleName}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testApp, ctx := CreateTestApp(true, t.TempDir())

			fromVM := testApp.mm.GetVersionMap()
			for _, name := range tc.missing {
				delete(fromVM, name)
			}

			handler := v089.CreateUpgradeHandler(
				testApp.mm,
				testApp.configurator,
				&testApp.AppKeepers,
				testApp.appCodec,
			)

			vm, err := handler(ctx, upgradetypes.Plan{Name: v089.UpgradeName}, fromVM)
			require.NoError(t, err, "the v0.8.9 handler must run without error")
			require.Equal(t, testApp.mm.GetVersionMap(), vm, "the handler must return every module at its current consensus version")
		})
	}
}
