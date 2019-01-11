package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/gorilla/mux"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutesQ(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc(fmt.Sprintf("/distribution/validator-rewards/{%s}", "validator-addr"), queryDistInfoHandlerFn(cdc, cliCtx)).Methods("GET")
	r.HandleFunc("/distribution/feepool", queryFeePoolHandlerFn(cdc, cliCtx)).Methods("GET")
}

func queryDistInfoHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		bech32ValAddr := vars["validator-addr"]

		if len(bech32ValAddr) == 0 {
			err := errors.New("validator-address required but not specified")
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		valAddr, err := sdk.ValAddressFromBech32(bech32ValAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := cliCtx.QueryStore(cmn.HexBytes(distribution.GetValidatorDistInfoKey(valAddr)), "distr")
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var vdi types.ValidatorDistInfo
		cdc.MustUnmarshalBinaryLengthPrefixed(res, &vdi)

		utils.PostProcessResponse(w, cdc, vdi, cliCtx.Indent)
	}
}

func queryFeePoolHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		res, err := cliCtx.QueryStore(cmn.HexBytes(distribution.FeePoolKey), "distr")
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var feepool types.FeePool
		cdc.MustUnmarshalBinaryLengthPrefixed(res, &feepool)

		utils.PostProcessResponse(w, cdc, feepool, cliCtx.Indent)
	}
}
