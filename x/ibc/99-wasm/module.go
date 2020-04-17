package wasm

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/99-wasm/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/99-wasm/client/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/99-wasm/types"
)

// Name returns the IBC client name
func Name() string {
	return types.SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the IBC client
func RegisterRESTRoutes(ctx client.Context, rtr *mux.Router, queryRoute string) {
	rest.RegisterRoutes(ctx, rtr, fmt.Sprintf("%s/%s", queryRoute, types.SubModuleName))
}

// GetTxCmd returns the root tx command for the IBC client
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	return cli.GetTxCmd(cdc, fmt.Sprintf("%s/%s", storeKey, types.SubModuleName))
}

// GetQueryCmd returns the root tx command for the IBC client
func GetQueryCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	return cli.GetQueryCmd(cdc, fmt.Sprintf("%s/%s", storeKey, types.SubModuleName))
}
