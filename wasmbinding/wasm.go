package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	contractmanagerkeeper "github.com/neutron-org/neutron/v2/x/contractmanager/keeper"
	feerefunderkeeper "github.com/neutron-org/neutron/v2/x/feerefunder/keeper"
	interchainqueriesmodulekeeper "github.com/neutron-org/neutron/v2/x/interchainqueries/keeper"
	interchaintransactionsmodulekeeper "github.com/neutron-org/neutron/v2/x/interchaintxs/keeper"
	transfer "github.com/neutron-org/neutron/v2/x/transfer/keeper"
)

// RegisterCustomPlugins returns wasmkeeper.Option that we can use to connect handlers for implemented custom queries and messages to the App.
func RegisterCustomPlugins(
	ictxKeeper *interchaintransactionsmodulekeeper.Keeper,
	icqKeeper *interchainqueriesmodulekeeper.Keeper,
	transfer transfer.KeeperTransferWrapper,
	feeRefunderKeeper *feerefunderkeeper.Keeper,
	contractmanagerKeeper *contractmanagerkeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(ictxKeeper, icqKeeper, feeRefunderKeeper, contractmanagerKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messageHandlerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(ictxKeeper, icqKeeper, transfer, contractmanagerKeeper),
	)

	return []wasmkeeper.Option{
		queryPluginOpt,
		messageHandlerDecoratorOpt,
	}
}
