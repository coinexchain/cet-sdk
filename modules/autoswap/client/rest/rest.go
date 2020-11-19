package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	mktrest "github.com/coinexchain/cet-sdk/modules/market/client/rest"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	// market routes
	mktrest.RegisterRoutes(cliCtx, r, cdc)
	// new routes
	registerTxRoutes(cliCtx, r, cdc)
	registerQueryRoutes(cliCtx, r, cdc)
}

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/market/add-liquidity", addLiquidityHandlerFn(cdc, cliCtx)).Methods("POST")
	r.HandleFunc("/market/remove-liquidity", removeLiquidityHandlerFn(cdc, cliCtx)).Methods("POST")
}

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/market/pool-list", queryPoolListHandlerFn(cdc, cliCtx)).Methods("GET")
	r.HandleFunc("/market/pool-info", queryPoolInfoHandlerFn(cdc, cliCtx)).Methods("GET")
}
