package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/dao"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"strings"

	"github.com/spf13/cobra"

	daoClientUtils "github.com/cosmos/cosmos-sdk/x/dao/client/utils"
)

const (
	flagTitle        = "title"
	flagDescription  = "description"
	flagProposalType = "type"
	flagDeposit      = "deposit"
	flagVoter        = "voter"
	flagOption       = "option"
	flagDepositor    = "depositor"
	flagStatus       = "status"
	flagNumLimit     = "limit"
	flagProposal     = "proposal"
	flagDenom = "denom"
	flagAmount = "amount"
)

type proposal struct {
	Title       string
	Description string
	Denom string
	Type        string
	Deposit     string
}

var proposalFlags = []string{
	flagTitle,
	flagDescription,
	flagProposalType,
	flagDeposit,
}

// TODO: SubmitProposal DAO must contain redelegation object, so need to broadcast json not cli
// GetCmdSubmitProposalDao implements submitting a proposal transaction command.
func GetCmdSubmitProposalDao(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal",
		Short: "Submit a proposal along with an initial deposit",
		Long: strings.TrimSpace(`
Submit a proposal along with an initial deposit. Proposal title, description, type and deposit can be given directly or through a proposal JSON file. For example:

$ gaiacli gov submit-proposal --proposal="path/to/proposal.json" --from mykey

where proposal.json contains:

{
  "title": "Test Proposal",
  "description": "My awesome proposal",
  "denom": "index1",
  "type": "Text",
  "deposit": "10test"
}

is equivalent to

$ gaiacli gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10test" --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposal, err := parseSubmitProposalFlags()
			if err != nil {
				return err
			}

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Get from address
			from := cliCtx.GetFromAddress()

			// Pull associated account
			account, err := cliCtx.GetAccount(from)
			if err != nil {
				return err
			}

			// Find deposit amount
			amount, err := sdk.ParseCoins(proposal.Deposit)
			if err != nil {
				return err
			}

			// ensure account has enough coins
			if !account.GetCoins().IsAllGTE(amount) {
				return fmt.Errorf("address %s doesn't have enough coins to pay for this transaction", from)
			}

			proposalType, err := dao.ProposalTypeFromString(proposal.Type)
			if err != nil {
				return err
			}

			msg := dao.NewMsgSubmitProposal(proposal.Title, proposal.Description, denom, proposalType, from, amount, dao.Rebalancing{})
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}

	cmd.Flags().String(flagTitle, "", "title of proposal")
	cmd.Flags().String(flagDescription, "", "description of proposal")
	cmd.Flags().String(flagProposalType, "", "proposalType of proposal, types: text/parameter_change")
	cmd.Flags().String(flagDeposit, "", "deposit of proposal")
	cmd.Flags().String(flagProposal, "", "proposal file path (if this path is given, other proposal flags are ignored)")

	return cmd
}

// GetCmdDepositDao implements depositing tokens for an active proposal.
func GetCmdDepositDao(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [proposal-id] [deposit]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit tokens for activing proposal",
		Long: strings.TrimSpace(`
Submit a deposit for an acive proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli tx gov deposit 1 10stake --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = daoClientUtils.QueryProposalByID(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			from := cliCtx.GetFromAddress()

			// Fetch associated account
			account, err := cliCtx.GetAccount(from)
			if err != nil {
				return err
			}

			// Get amount of coins
			amount, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			// ensure account has enough coins
			if !account.GetCoins().IsAllGTE(amount) {
				return fmt.Errorf("address %s doesn't have enough coins to pay for this transaction", from)
			}

			msg := dao.NewMsgDeposit(from, proposalID, amount)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

// GetCmdVoteDao implements creating a new vote command.
func GetCmdVoteDao(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "vote [proposal-id] [option] [amount]",
		Args:  cobra.ExactArgs(3),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain, amount: amount of the proposal Denom for voting, it will be Staked in voting period",
		Long: strings.TrimSpace(`
Submit a vote for an acive proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli tx gov vote 1 yes --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Get voting address
			from := cliCtx.GetFromAddress()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = daoClientUtils.QueryProposalByID(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Find out which vote option user chose
			byteVoteOption, err := dao.VoteOptionFromString(daoClientUtils.NormalizeVoteOption(args[1]))
			if err != nil {
				return err
			}

			// parse coins trying to be sent
			coin, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			// Pull associated account
			account, err := cliCtx.GetAccount(from)
			if err != nil {
				return err
			}

			// ensure account has enough coins
			if !account.GetCoins().IsAllGTE(sdk.Coins{coin}) {
				return fmt.Errorf("address %s doesn't have enough coins to pay for this transaction", from)
			}

			// Build vote message and run basic validation
			msg := dao.NewMsgVoteDao(from, proposalID, byteVoteOption, coin)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

// DONTCOVER
