package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/tendermint/spm/cosmoscmd"
	tmcmds "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/app"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/app/params"
)

func main() {
	params.SetAddressPrefixes()
	cmdOptions := GetWasmCmdOptions()
	cmdOptions = append(cmdOptions, cosmoscmd.AddSubCmd(tmcmds.RollbackStateCmd))
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
