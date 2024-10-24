package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	contractmanagerkeeper "github.com/Nolus-Protocol/nolus-core/x/contractmanager/keeper"
	feerefunderkeeper "github.com/Nolus-Protocol/nolus-core/x/feerefunder/keeper"
	interchaintransactionsmodulekeeper "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/keeper"
	transfer "github.com/Nolus-Protocol/nolus-core/x/transfer/keeper"
)

// RegisterCustomPlugins returns wasmkeeper.Option that we can use to connect handlers for implemented custom queries and messages to the App.
func RegisterCustomPlugins(
	ictxKeeper *interchaintransactionsmodulekeeper.Keeper,
	transfer transfer.KeeperTransferWrapper,
	feeRefunderKeeper *feerefunderkeeper.Keeper,
	contractmanagerKeeper *contractmanagerkeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(ictxKeeper, feeRefunderKeeper, contractmanagerKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messageHandlerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(ictxKeeper, transfer, contractmanagerKeeper),
	)

	return []wasmkeeper.Option{
		queryPluginOpt,
		messageHandlerDecoratorOpt,
	}
}
