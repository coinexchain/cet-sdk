package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/coinexchain/cet-sdk/modules/incentive/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/incentive/internal/types"
	"github.com/coinexchain/cosmos-utils/client/restutil"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/incentive/parameters", queryParamsHandlerFn(cliCtx)).Methods("GET")
}

// HTTP request handler to query the alias params values
func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.StoreKey, keepers.QueryParameters)
		restutil.RestQuery(nil, cliCtx, w, r, route, nil, nil)
	}
}
