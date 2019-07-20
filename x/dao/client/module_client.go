package client

import (
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/dao"
	govCli "github.com/cosmos/cosmos-sdk/x/dao/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:   dao.ModuleName,
		Short: "Querying commands for the governance module",
	}

	govQueryCmd.AddCommand(client.GetCommands(
		govCli.GetCmdQueryProposal(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryProposals(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryVote(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryVotes(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryParam(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryParams(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryProposer(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryDeposit(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryDeposits(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryTally(mc.storeKey, mc.cdc))...)

	return govQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:   gov.ModuleName,
		Short: "Governance transactions subcommands",
	}

	govTxCmd.AddCommand(client.PostCommands(
		govCli.GetCmdDepositDao(mc.storeKey, mc.cdc),
		govCli.GetCmdVoteDao(mc.storeKey, mc.cdc),
		govCli.GetCmdSubmitProposalDao(mc.cdc),
	)...)

	return govTxCmd
}
