package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/gorilla/mux"

	"github.com/coinexchain/cet-sdk/modules/authx/internal/types"
	"github.com/coinexchain/cosmos-utils/client/restutil"
)

// register REST routes
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/auth/accounts/{address}", QueryAccountRequestHandlerFn(cliCtx, cdc)).Methods("GET")
	r.HandleFunc("/auth/parameters", QueryParamsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/auth/accounts/{address}/setReferee", setRefereeHandleFn(cdc, cliCtx)).Methods("POST")
}

// query accountREST Handler
func QueryAccountRequestHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.StoreKey, types.QueryAccountMix)
		vars := mux.Vars(r)
		acc, err := sdk.AccAddressFromBech32(vars["address"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		params := auth.NewQueryAccountParams(acc)

		restutil.RestQuery(cdc, cliCtx, w, r, route, &params, nil)
	}
}

// HTTP request handler to query the authx params values
func QueryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.StoreKey, types.QueryParameters)
		restutil.RestQuery(nil, cliCtx, w, r, route, nil, nil)
	}
}
