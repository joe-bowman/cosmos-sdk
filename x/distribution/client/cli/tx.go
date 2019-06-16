// nolint
package cli

import (
	"strings"

	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	flagOnlyFromValidator = "only-from-validator"
	flagIsValidator       = "is-validator"
	flagComission         = "commission"
	flagMaxMessagesPerTx  = "max-msgs"
)

const (
	MaxMessagesPerTxDefault = 5
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *amino.Codec) *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:   "dist",
		Short: "Distribution transactions subcommands",
	}

	distTxCmd.AddCommand(client.PostCommands(
		GetCmdSetWithdrawAddr(cdc),
	)...)

	return distTxCmd
}

type generateOrBroadcastFunc func(context.CLIContext, authtxb.TxBuilder, []sdk.Msg, bool) error

func splitAndApply(
	generateOrBroadcast generateOrBroadcastFunc,
	cliCtx context.CLIContext,
	txBldr authtxb.TxBuilder,
	msgs []sdk.Msg,
	chunkSize int,
	offline bool,
) error {

	if chunkSize == 0 {
		return generateOrBroadcast(cliCtx, txBldr, msgs, offline)
	}

	// split messages into slices of length chunkSize
	totalMessages := len(msgs)
	for i := 0; i < len(msgs); i += chunkSize {

		sliceEnd := i + chunkSize
		if sliceEnd > totalMessages {
			sliceEnd = totalMessages
		}

		msgChunk := msgs[i:sliceEnd]
		if err := generateOrBroadcast(cliCtx, txBldr, msgChunk, offline); err != nil {
			return err
		}
	}

	return nil
}

// command to withdraw rewards
func GetCmdWithdrawCommission(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-commission",
		Short: "withdraw validators comission for the given validator",
		Long: strings.TrimSpace(`withdraw validators comission for the given validator:

$ gaiacli tx distr withdraw-commission --from mykey
`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			delAddr := cliCtx.GetFromAddress()
			valAddr := sdk.ValAddress(delAddr)

			msgs := []sdk.Msg{types.NewMsgWithdrawValidatorCommission(valAddr)}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, msgs, false)
		},
	}
	cmd.Flags().Bool(flagComission, false, "also withdraw validator's commission")
	return cmd
}

// command to withdraw all rewards
// func GetCmdWithdrawAllRewards(cdc *codec.Codec, queryRoute string) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "withdraw-all-rewards",
// 		Short: "withdraw all delegations rewards for a delegator",
// 		Long: strings.TrimSpace(`Withdraw all rewards for a single delegator:
//
// $ gaiacli tx distr withdraw-all-rewards --from mykey
// `),
// 		Args: cobra.NoArgs,
// 		RunE: func(cmd *cobra.Command, args []string) error {
//
// 			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
// 			cliCtx := context.NewCLIContext().
// 				WithCodec(cdc).
// 				WithAccountDecoder(cdc)
//
// 			delAddr := cliCtx.GetFromAddress()
// 			msgs, err := common.WithdrawAllDelegatorRewards(cliCtx, cdc, queryRoute, delAddr)
// 			if err != nil {
// 				return err
// 			}
//
// 			chunkSize := viper.GetInt(flagMaxMessagesPerTx)
// 			return splitAndApply(utils.GenerateOrBroadcastMsgs, cliCtx, txBldr, msgs, chunkSize, false)
// 		},
// 	}
//
// 	cmd.Flags().Int(flagMaxMessagesPerTx, MaxMessagesPerTxDefault, "Limit the number of messages per tx (0 for unlimited)")
// 	return cmd
// }

// command to replace a delegator's withdrawal address
func GetCmdSetWithdrawAddr(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-withdraw-addr [withdraw-addr]",
		Short: "change the default withdraw address for rewards associated with an address",
		Long: strings.TrimSpace(`Set the withdraw address for rewards associated with a delegator address:

$ gaiacli tx set-withdraw-addr cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p --from mykey
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			delAddr := cliCtx.GetFromAddress()
			withdrawAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}
