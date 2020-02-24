package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/coinexchain/cet-sdk/modules/bancorlite/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/bancorlite/internal/types"
	"github.com/coinexchain/cet-sdk/modules/market"
	"github.com/coinexchain/cosmos-utils/client/restutil"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/bancorlite/pools/{symbol}", queryBancorInfoHandlerFn(cdc, cliCtx)).Methods("GET")
	r.HandleFunc("/bancorlite/parameters", queryParamsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/bancorlite/infos", queryBancorsHandlerFn(cdc, cliCtx)).Methods("GET")
}

// format: barcorlite/pools/btc-cet
func queryBancorInfoHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		query := fmt.Sprintf("custom/%s/%s", types.StoreKey, keepers.QueryBancorInfo)
		symbol := strings.Replace(vars["symbol"], "-", "/", 1)
		if !market.IsValidTradingPair(strings.Split(symbol, "/")) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "Invalid Trading pair")
			return
		}
		param := &keepers.QueryBancorInfoParam{Symbol: symbol}
		restutil.RestQuery(cdc, cliCtx, w, r, query, param, nil)
	}
}

func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.StoreKey, keepers.QueryParameters)
		restutil.RestQuery(nil, cliCtx, w, r, route, nil, nil)
	}
}

func queryBancorsHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := fmt.Sprintf("custom/%s/%s", types.StoreKey, keepers.QueryBancors)
		restutil.RestQuery(cdc, cliCtx, w, r, query, nil, nil)
	}
}
