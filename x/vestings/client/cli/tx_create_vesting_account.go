package cli

import (
	"errors"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/x/vestings/types"
)

var _ = strconv.Itoa(0)

// Transaction command flags.
const (
	FlagDelayed = "delayed"
)

func CmdCreateVestingAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-vesting-account [to_address] [amount] [start_time] [end_time]",
		Short: "Create a new vesting account funded with an allocation of tokens.",
		Long: `Create a new vesting account funded with an allocation of tokens. The
account can either be a delayed or continuous vesting account, which is determined
by the '--delayed' flag. The start_time, end_time must be provided as a UNIX epoch
timestamp.`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			if args[1] == "" {
				return errors.New("amount is empty")
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			startTime, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			endTime, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			delayed, err := cmd.Flags().GetBool(FlagDelayed)
			if err != nil {
				return err
			}
			msg := types.NewMsgCreateVestingAccount(
				clientCtx.GetFromAddress(),
				toAddr,
				amount,
				startTime,
				endTime,
				delayed,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().Bool(FlagDelayed, false, "Create a delayed vesting account if true")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
