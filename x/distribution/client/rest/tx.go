package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
  "github.com/cosmos/cosmos-sdk/crypto/keys"
  sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	r.HandleFunc("/distribution/withdraw-validator-rewards", withdrawHandlerFn(cdc, kb, cliCtx)).Methods("POST")
	r.HandleFunc("/distribution/set-withdraw-address", setWithdrawAddressHandlerFn(cdc, kb, cliCtx)).Methods("POST")
  r.HandleFunc("/distribution/withdraw-delegator-rewards", withdrawHandlerDelegatorFn(cdc, kb, cliCtx)).Methods("POST")
}

type WithdrawDelegatorReq struct {
	BaseReq           utils.BaseReq    `json:"base_req"`
  OnlyFromValidator string           `json:"only_from_validator"`
}

type WithdrawValidatorReq struct {
	BaseReq      utils.BaseReq  `json:"base_req"`
}

type SetWithdrawAddressReq struct {
	BaseReq      utils.BaseReq  `json:"base_req"`
	WithdrawAddr string         `json:"withdraw_addr"` // bech32 Address of the account to withdraw to
}

func withdrawHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {

    var req WithdrawValidatorReq
    err := utils.ReadRESTReq(w, r, cdc, &req)
    if err != nil {
      return
    }

    baseReq := req.BaseReq.Sanitize()
    if !baseReq.ValidateBasic(w, cliCtx) {
      return
    }

    info, err := kb.Get(baseReq.Name)
    if err != nil {
      utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
      return
    }

    //var msg sdk.Msg

    valAddr := sdk.ValAddress(info.GetAddress())
    msg := types.NewMsgWithdrawValidatorRewardsAll(valAddr)

    // build and sign the transaction, then broadcast to Tendermint
    utils.CompleteAndBroadcastTxREST(w, r, cliCtx, baseReq, []sdk.Msg{msg}, cdc)
  }
}

func withdrawHandlerDelegatorFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    var req WithdrawDelegatorReq

    err := utils.ReadRESTReq(w, r, cdc, &req)
    if err != nil {
      return
    }

    baseReq := req.BaseReq.Sanitize()
    if !baseReq.ValidateBasic(w, cliCtx) {
      return
    }

    info, err := kb.Get(baseReq.Name)
    if err != nil {
      utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
      return
    }

    delAddr := sdk.AccAddress(info.GetAddress())

    var msg sdk.Msg
    switch {
    case req.OnlyFromValidator != "":

      valAddr, err := sdk.ValAddressFromBech32(req.OnlyFromValidator)
      if err != nil {
        utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
      }
      msg = types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
    default:
      msg = types.NewMsgWithdrawDelegatorRewardsAll(delAddr)
    }

    // build and sign the transaction, then broadcast to Tendermint
		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, baseReq, []sdk.Msg{msg}, cdc)
  }
}



func setWithdrawAddressHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    var req SetWithdrawAddressReq
    err := utils.ReadRESTReq(w, r, cdc, &req)
    if err != nil {
      return
    }

    baseReq := req.BaseReq.Sanitize()
    if !baseReq.ValidateBasic(w, cliCtx) {
      return
    }

    info, err := kb.Get(baseReq.Name)
    if err != nil {
      utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
      return
    }

    valAddr := info.GetAddress()

    withdrawAddr := valAddr

    bech32withdraw := req.WithdrawAddr
    if bech32withdraw != "" {
      withdrawAddr, err = sdk.AccAddressFromBech32(bech32withdraw)
      if err != nil {
        utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
      }
    }

    msg := types.NewMsgSetWithdrawAddress(withdrawAddr, valAddr)

    // build and sign the transaction, then broadcast to Tendermint
    utils.CompleteAndBroadcastTxREST(w, r, cliCtx, baseReq, []sdk.Msg{msg}, cdc)
  }
}
