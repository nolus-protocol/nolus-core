package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"

	"github.com/cosmos/cosmos-sdk/client"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdChangeSuspend())
// this line is used by starport scaffolding # 1

	return cmd
}
