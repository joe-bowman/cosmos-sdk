package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// GetCmdQueryParams implements the query params command.
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query distribution params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			params, err := common.QueryParams(cliCtx, queryRoute)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(params)
		},
	}
}

// GetCmdQueryValidatorCommission implements the query validator commission command.
func GetCmdQueryValidatorCommission(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "commission [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query distribution validator commission",
		Long: strings.TrimSpace(`Query validator commission rewards from delegators to that validator:

$ gaiacli query distr commission cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := common.QueryValidatorCommission(cliCtx, cdc, queryRoute, validatorAddr)
			if err != nil {
				return err
			}

			var valCom types.ValidatorAccumulatedCommission
			cdc.MustUnmarshalJSON(res, &valCom)
			return cliCtx.PrintOutput(valCom)
		},
	}
}

// GetCmdQueryValidatorSlashes implements the query validator slashes command.
func GetCmdQueryValidatorSlashes(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "slashes [validator] [start-height] [end-height]",
		Args:  cobra.ExactArgs(3),
		Short: "Query distribution validator slashes",
		Long: strings.TrimSpace(`Query all slashes of a validator for a given block range:

$ gaiacli query distr slashes cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 0 100
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			startHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("start-height %s not a valid uint, please input a valid start-height", args[1])
			}

			endHeight, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("end-height %s not a valid uint, please input a valid end-height", args[2])
			}

			params := distr.NewQueryValidatorSlashesParams(validatorAddr, startHeight, endHeight)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validator_slashes", queryRoute), bz)
			if err != nil {
				return err
			}

			var slashes types.ValidatorSlashEvents
			cdc.MustUnmarshalJSON(res, &slashes)
			return cliCtx.PrintOutput(slashes)
		},
	}
}

// GetCmdQueryCommunityPool returns the command for fetching community pool info
func GetCmdQueryCommunityPool(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "community-pool",
		Args:  cobra.NoArgs,
		Short: "Query the amount of coins in the community pool",
		Long: strings.TrimSpace(`Query all coins in the community pool which is under Governance control.

$ gaiacli query distr community-pool
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/community_pool", queryRoute), nil)
			if err != nil {
				return err
			}

			var result sdk.DecCoins
			cdc.MustUnmarshalJSON(res, &result)
			return cliCtx.PrintOutput(result)
		},
	}
}
