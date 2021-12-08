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
		Use:   "change-suspend [suspend]",
		Short: "Broadcast message change-suspend",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSuspend, err := cast.ToBoolE(args[0])
			argAdminKey, err := cast.ToStringE(args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgChangeSuspend(
				clientCtx.GetFromAddress().String(),
				argSuspend,
				argAdminKey,
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
