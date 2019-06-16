package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router,
	cdc *codec.Codec, queryRoute string) {

	// Get the total rewards balance from all delegations
	// r.HandleFunc(
	// 	"/distribution/delegators/{delegatorAddr}/rewards",
	// 	delegatorRewardsHandlerFn(cliCtx, cdc, queryRoute),
	// ).Methods("GET")

	// Query a delegation reward
	// r.HandleFunc(
	// 	"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
	// 	delegationRewardsHandlerFn(cliCtx, cdc, queryRoute),
	// ).Methods("GET")

	// Get the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		delegatorWithdrawalAddrHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Validator distribution information
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}",
		validatorInfoHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Get the current distribution parameter values
	r.HandleFunc(
		"/distribution/parameters",
		paramsHandlerFn(cliCtx, cdc, queryRoute),
	).Methods("GET")

	// Get the amount held in the community pool
	r.HandleFunc(
		"/distribution/community_pool",
		communityPoolHandler(cliCtx, cdc, queryRoute),
	).Methods("GET")

}

// HTTP request handler to query a delegation rewards
func delegatorWithdrawalAddrHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		delegatorAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		bz := cdc.MustMarshalJSON(distribution.NewQueryDelegatorWithdrawAddrParams(delegatorAddr))
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/withdraw_addr", queryRoute), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// ValidatorDistInfo defines the properties of
// validator distribution information response.
type ValidatorDistInfo struct {
	OperatorAddress     sdk.AccAddress                       `json:"operator_address"`
	SelfBondRewards     sdk.DecCoins                         `json:"self_bond_rewards"`
	ValidatorCommission types.ValidatorAccumulatedCommission `json:"val_commission"`
}

// NewValidatorDistInfo creates a new instance of ValidatorDistInfo.
func NewValidatorDistInfo(operatorAddr sdk.AccAddress, rewards sdk.DecCoins,
	commission types.ValidatorAccumulatedCommission) ValidatorDistInfo {
	return ValidatorDistInfo{
		OperatorAddress:     operatorAddr,
		SelfBondRewards:     rewards,
		ValidatorCommission: commission,
	}
}

// HTTP request handler to query validator's distribution information
func validatorInfoHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		//valAddr := mux.Vars(r)["validatorAddr"]
		validatorAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		// query commission
		commissionRes, err := common.QueryValidatorCommission(cliCtx, cdc, queryRoute, validatorAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var valCom types.ValidatorAccumulatedCommission
		cdc.MustUnmarshalJSON(commissionRes, &valCom)

		// self bond rewards
		delAddr := sdk.AccAddress(validatorAddr)

		// Prepare response
		res := cdc.MustMarshalJSON(NewValidatorDistInfo(delAddr, sdk.DecCoins{}, valCom))
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// HTTP request handler to query the distribution params values
func paramsHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		params, err := common.QueryParams(cliCtx, queryRoute)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cdc, params, cliCtx.Indent)
	}
}

func communityPoolHandler(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/community_pool", queryRoute), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var result sdk.DecCoins
		if err := cdc.UnmarshalJSON(res, &result); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, result, cliCtx.Indent)
	}
}
