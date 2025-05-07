package transfer

import (
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	wrapkeeper "github.com/Nolus-Protocol/nolus-core/x/transfer/keeper"
	transfertypes "github.com/Nolus-Protocol/nolus-core/x/transfer/types"
)

/*
	In addition to original ack processing of ibc transfer acknowledgement we want to pass the acknowledgement to originating wasm contract.
	The package contains a code to achieve the purpose.
*/

type IBCModule struct {
	wrappedKeeper wrapkeeper.KeeperTransferWrapper
	keeper        keeper.Keeper
	sudoKeeper    transfertypes.WasmKeeper
	transfer.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper.
func NewIBCModule(k wrapkeeper.KeeperTransferWrapper, sudoKeeper transfertypes.WasmKeeper) IBCModule {
	return IBCModule{
		wrappedKeeper: k,
		keeper:        k.Keeper,
		sudoKeeper:    sudoKeeper,
		IBCModule:     transfer.NewIBCModule(k.Keeper),
	}
}

// OnAcknowledgementPacket implements the IBCModule interface.
// Wrapper struct shadows(overrides) the OnAcknowledgementPacket method to achieve the package's purpose.
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	err := im.IBCModule.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	if err != nil {
		return errors.Wrap(err, "failed to process original OnAcknowledgementPacket")
	}
	return im.HandleAcknowledgement(ctx, channelVersion, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	err := im.IBCModule.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	if err != nil {
		return errors.Wrap(err, "failed to process original OnTimeoutPacket")
	}
	return im.HandleTimeout(ctx, channelVersion, packet, relayer)
}

var _ appmodule.AppModule = AppModule{}

type AppModule struct {
	transfer.AppModule
	keeper wrapkeeper.KeeperTransferWrapper
}

// NewAppModule creates a new 20-transfer module.
func NewAppModule(k wrapkeeper.KeeperTransferWrapper) AppModule {
	return AppModule{
		AppModule: transfer.NewAppModule(k.Keeper),
		keeper:    k,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() { // marker
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() { // marker
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
	transfertypes.RegisterMsgServer(cfg.MsgServer(), am.keeper)

	cfg.MsgServer().RegisterService(&transfertypes.MsgServiceDescOrig, am.keeper)

	m := keeper.NewMigrator(am.keeper.Keeper)
	if err := cfg.RegisterMigration(types.ModuleName, 2, m.MigrateTotalEscrowForDenom); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 2 to 3: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 3, m.MigrateParams); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app version 3 to 4: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 4, m.MigrateDenomMetadata); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 4 to 5: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 5, m.MigrateDenomTraceToDenom); err != nil {
		panic(fmt.Errorf("failed to migrate transfer app from version 5 to 6 (migrate DenomTrace to Denom): %v", err))
	}
}

type AppModuleBasic struct {
	transfer.AppModuleBasic
}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{AppModuleBasic: transfer.AppModuleBasic{}}
}

func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	transfertypes.RegisterLegacyAminoCodec(cdc)
}

func (am AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	transfertypes.RegisterLegacyAminoCodec(cdc)
	am.AppModuleBasic.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types.
func (am AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	transfertypes.RegisterInterfaces(reg)
	am.AppModuleBasic.RegisterInterfaces(reg)
}

// Name returns the interchaintxs module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}
