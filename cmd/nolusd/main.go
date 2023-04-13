package main

import (
	"os"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/app/params"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/tendermint/spm/cosmoscmd"
)

func main() {
	// we must override the wasm variables here because we want to upload contracts on genesis
	// and in our scripts, we use the cli command add-wasm-genesis-message before the chain is started in order to load the contracts
	overrideWasmVariables()

	params.SetAddressPrefixes()
	cmdOptions := GetWasmCmdOptions()
	rootCmd, _ := cosmoscmd.NewRootCmd(
		app.Name,
		params.Bech32PrefixAccAddr,
		app.DefaultNodeHome,
		app.Name,
		app.ModuleBasics,
		app.New,
		cmdOptions...,
	)

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}

// overrideWasmVariables overrides the wasm variables to:
//   - allow for larger wasm files
func overrideWasmVariables() {
	// Override Wasm size limitation from WASMD.
	wasmtypes.MaxWasmSize = 3 * 1024 * 1024
	wasmtypes.MaxProposalWasmSize = wasmtypes.MaxWasmSize
}
