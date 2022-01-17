package cli

import (
	"strconv"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

var _ = strconv.Itoa(0)

func CmdChangeSuspend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-suspend [suspended] [block-height]",
		Short: "Broadcast message change-suspend",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSuspend, err := cast.ToBoolE(args[0])
			if err != nil {
				return err
			}
			argBlockHeight, err := cast.ToInt64E(args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSuspend(
				clientCtx.GetFromAddress().String(),
				argSuspend,
				argBlockHeight,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
