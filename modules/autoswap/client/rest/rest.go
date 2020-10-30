package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	registerTxRoutes(cliCtx, r, cdc)
	registerQueryRoutes(cliCtx, r, cdc)
}

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/autoswap/add-liquidity", addLiquidityHandlerFn(cdc, cliCtx)).Methods("POST")
	r.HandleFunc("/autoswap/remove-liquidity", removeLiquidityHandlerFn(cdc, cliCtx)).Methods("POST")
	r.HandleFunc("/autoswap/market-orders", createMarketOrderHandlerFn(cdc, cliCtx)).Methods("POST")
	r.HandleFunc("/autoswap/limit-orders", createLimitOrderHandlerFn(cdc, cliCtx)).Methods("POST")
	r.HandleFunc("/autoswap/cancel-order", cancelOrderHandlerFn(cdc, cliCtx)).Methods("POST")
}

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/autoswap/parameters", queryParamsHandlerFn(cliCtx)).Methods("GET")
	//r.HandleFunc("/autoswap/pools/{stock}/{money}", queryMarketHandlerFn(cdc, cliCtx)).Methods("GET")
	//r.HandleFunc("/autoswap/orderbook/{stock}/{money}", queryOrdersInMarketHandlerFn(cdc, cliCtx)).Methods("GET")
	//r.HandleFunc("/autoswap/exist-trading-pairs", queryMarketsHandlerFn(cdc, cliCtx)).Methods("GET")
	//r.HandleFunc("/autoswap/orders/{order-id}", queryOrderInfoHandlerFn(cdc, cliCtx)).Methods("GET")
	//r.HandleFunc("/autoswap/orders/account/{address}", queryUserOrderListHandlerFn(cdc, cliCtx)).Methods("GET")
}
