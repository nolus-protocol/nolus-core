package main

import (
	"os"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/app/params"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	config := params.GetDefaultConfig()
	config.Seal()
	// we must override the wasm variables here because we want to upload contracts on genesis
	// and in our scripts, we use the cli command add-wasm-genesis-message before the chain is started in order to load the contracts
	overrideWasmVariables()

	rootCmd, _ := NewRootCmd(
		app.Name,
		app.DefaultNodeHome,
		app.Name,
	)

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}

// overrideWasmVariables overrides the wasm variables to:
//   - allow for larger wasm files
func overrideWasmVariables() {
	// Override Wasm size limitation from WASMD.
	wasmtypes.MaxWasmSize = 5 * 1024 * 1024
	wasmtypes.MaxProposalWasmSize = wasmtypes.MaxWasmSize
}
