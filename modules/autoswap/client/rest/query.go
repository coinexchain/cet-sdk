package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/market"
	dex "github.com/coinexchain/cet-sdk/types"
	"github.com/coinexchain/cosmos-utils/client/restutil"
)

func queryPoolListHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := fmt.Sprintf("custom/%s/%s", market.StoreKey, keepers.QueryPools)
		restutil.RestQuery(cdc, cliCtx, w, r, query, nil, nil)
	}
}

func queryPoolInfoHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		route := fmt.Sprintf("custom/%s/%s", market.StoreKey, keepers.QueryPoolInfo)
		if !market.IsValidTradingPair([]string{vars["stock"], vars["money"]}) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "Invalid Trading pair")
			return
		}
		param := market.QueryMarketParam{TradingPair: dex.GetSymbol(vars["stock"], vars["money"])}
		restutil.RestQuery(cdc, cliCtx, w, r, route, param, nil)
	}
}
