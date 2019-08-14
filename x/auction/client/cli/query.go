package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

type StringyBytes []byte

func (s StringyBytes) String() string { return string(s) }

// GetCmdGetAuctions queries the auctions in the store
func GetCmdGetAuctions(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "getauctions",
		Short: "get a list of active auctions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getauctions", queryRoute), nil)
			if err != nil {
				fmt.Printf("error when getting auctions - %s", err)
				return nil
			}

			return cliCtx.PrintOutput(StringyBytes(res))
		},
	}
}
