package app

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	v086 "github.com/Nolus-Protocol/nolus-core/app/upgrades/v086"
	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// The module manager panics at construction when a module implementing a genesis
// or blocker interface is missing from the matching SetOrder* list, so building
// the app at all is most of this assertion; the explicit checks pin down which
// lists the module genuinely belongs to.
func TestSolanaCarrierIsWiredIntoTheApp(t *testing.T) {
	testApp, _ := CreateTestApp(true, t.TempDir())

	require.Contains(t, ModuleBasics, solanacarrier.ModuleName,
		"module must be in ModuleBasics so its codec, genesis and gateway routes register")
	require.Contains(t, testApp.mm.Modules, solanacarrier.ModuleName,
		"module must be in appModules so the module manager drives it")

	require.True(t, slices.Contains(testApp.mm.OrderInitGenesis, solanacarrier.ModuleName),
		"module must be in the InitGenesis order")
	require.True(t, slices.Contains(testApp.mm.OrderExportGenesis, solanacarrier.ModuleName),
		"module must be in the ExportGenesis order")

	require.Contains(t, testApp.GetKVStoreKeys(), solanacarrier.StoreKey,
		"module store key must be mounted")
	require.NotNil(t, testApp.SolanaCarrierKeeper)
}

// The module has no BeginBlocker, EndBlocker or PreBlocker, and no
// AppModuleSimulation. SetOrderBeginBlockers and SetOrderEndBlockers only assert
// over modules implementing those interfaces, so its absence from those lists is
// the correct wiring rather than an omission.
func TestSolanaCarrierIsAbsentFromBlockerAndSimulationOrders(t *testing.T) {
	testApp, _ := CreateTestApp(true, t.TempDir())

	require.False(t, slices.Contains(testApp.mm.OrderBeginBlockers, solanacarrier.ModuleName))
	require.False(t, slices.Contains(testApp.mm.OrderEndBlockers, solanacarrier.ModuleName))
	require.False(t, slices.Contains(testApp.mm.OrderPreBlockers, solanacarrier.ModuleName))

	for _, m := range testApp.sm.Modules {
		named, ok := m.(module.HasName)
		if !ok {
			continue
		}
		require.NotEqual(t, solanacarrier.ModuleName, named.Name(),
			"module implements no AppModuleSimulation and must stay out of the simulation manager")
	}
}

// RegisterGRPCGatewayRoutes panics on a bad route and runs on every node that
// serves the API, so a broken pattern would crash the node at startup.
func TestSolanaCarrierRegistersItsGatewayRouteWithoutPanicking(t *testing.T) {
	testApp, _ := CreateTestApp(true, t.TempDir())

	clientCtx := client.Context{}.
		WithCodec(testApp.appCodec).
		WithInterfaceRegistry(testApp.interfaceRegistry)

	require.NotPanics(t, func() {
		ModuleBasics[solanacarrier.ModuleName].RegisterGRPCGatewayRoutes(clientCtx, runtime.NewServeMux())
	})
}

func TestSolanaCarrierInitGenesisPersistsTheDefaultFeePayer(t *testing.T) {
	testApp, ctx := CreateTestApp(true, t.TempDir())

	genesisModule, ok := testApp.mm.Modules[solanacarrier.ModuleName].(module.HasGenesis)
	require.True(t, ok, "module must implement module.HasGenesis")

	genesisJSON := genesisModule.DefaultGenesis(testApp.appCodec)
	require.NoError(t, genesisModule.ValidateGenesis(testApp.appCodec, testApp.encodingConfig.TxConfig, genesisJSON))

	genesisModule.InitGenesis(ctx, testApp.appCodec, genesisJSON)

	feePayer, err := testApp.SolanaCarrierKeeper.FeePayer(ctx)
	require.NoError(t, err)
	require.Equal(t, solanacarrier.DefaultFeePayer, feePayer)

	require.JSONEq(t, string(genesisJSON), string(genesisModule.ExportGenesis(ctx, testApp.appCodec)))
}

// The v0.8.6 handler must leave the fee payer seeded on a chain whose store never
// held this module, which is the state every upgrading node starts from.
func TestV086UpgradeHandlerSeedsTheFeePayer(t *testing.T) {
	testApp, ctx := CreateTestApp(true, t.TempDir())

	_, err := testApp.SolanaCarrierKeeper.FeePayer(ctx)
	require.ErrorIs(t, err, solanacarrier.ErrParamsNotFound,
		"precondition: the module's params must be absent before the upgrade")

	fromVM := testApp.mm.GetVersionMap()
	delete(fromVM, solanacarrier.ModuleName)

	handler := v086.CreateUpgradeHandler(
		testApp.mm,
		testApp.configurator,
		&testApp.AppKeepers,
		testApp.appCodec,
	)

	vm, err := handler(ctx, upgradetypes.Plan{Name: v086.UpgradeName}, fromVM)
	require.NoError(t, err)
	require.Equal(t, uint64(1), vm[solanacarrier.ModuleName],
		"the upgrade must record the module at consensus version 1")

	feePayer, err := testApp.SolanaCarrierKeeper.FeePayer(ctx)
	require.NoError(t, err)
	require.Equal(t, solanacarrier.DefaultFeePayer, feePayer)
}

func TestV086AddsTheSolanaCarrierStore(t *testing.T) {
	require.Equal(t, []string{solanacarrier.StoreKey}, v086.Upgrade.StoreUpgrades.Added)
	require.Empty(t, v086.Upgrade.StoreUpgrades.Deleted)
	require.Equal(t, "v0.8.6", v086.UpgradeName)

	names := make([]string, 0, len(Upgrades))
	for _, u := range Upgrades {
		names = append(names, u.UpgradeName)
	}
	require.Contains(t, names, v086.UpgradeName, "v0.8.6 must be registered in the Upgrades list")
}
