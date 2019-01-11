package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/gorilla/mux"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/mint/inflation", queryInflationHandlerFn(cdc, cliCtx)).Methods("GET")
}

func queryInflationHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		res, err := cliCtx.QueryStore(cmn.HexBytes([]byte{0x00}), "mint")
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var minter mint.Minter
		cdc.MustUnmarshalBinaryLengthPrefixed(res, &minter)

		utils.PostProcessResponse(w, cdc, minter, cliCtx.Indent)
	}
}
