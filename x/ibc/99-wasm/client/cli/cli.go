package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetTxCmd returns the transaction commands for IBC clients
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	ics99WasmTxCmd := &cobra.Command{
		Use:                        "wasm",
		Short:                      "IBC Wasm transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics99WasmTxCmd.AddCommand(flags.PostCommands(
		GetCmdCreateClient(cdc),
		GetCmdUpdateClient(cdc),
		GetCmdSubmitMisbehaviour(cdc),
		GetCmdStoreWasm(cdc),
	)...)

	return ics99WasmTxCmd
}

func GetQueryCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	ics99WasmQueryCmd := &cobra.Command{
		Use:                        "wasm",
		Short:                      "IBC Wasm query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics99WasmQueryCmd.AddCommand(flags.PostCommands(
		GetCmdListCode(cdc),
		GetCmdListContractByCode(cdc),
		GetCmdQueryCode(cdc),
		GetCmdGetContractInfo(cdc),
		GetCmdGetContractState(cdc),
	)...)

	return ics99WasmQueryCmd
}
