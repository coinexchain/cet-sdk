package keepers

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	QueryMarket            = "market-info"
	QueryMarkets           = "market-list"
	QueryOrdersInMarket    = "orders-in-market"
	QueryUserOrders        = "user-order-list"
	QueryWaitCancelMarkets = "wait-cancel-markets"
	//QueryOrder             = "order-info"
	//QueryParameters        = "parameters"
)

func NewQuerier2(mk Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryParameters:
			return queryParameters2(ctx, mk)
		case QueryMarket:
			return queryMarket(ctx, req, mk)
		case QueryMarkets:
			return queryMarketList(ctx, req, mk)
		case QueryOrdersInMarket:
			return queryOrdersInMarket(ctx, req, mk)
		case QueryOrder:
			return queryOrder(ctx, req, mk)
		case QueryUserOrders:
			return queryUserOrderList(ctx, req, mk)
		case QueryWaitCancelMarkets:
			return queryWaitCancelMarkets(ctx, req, mk)
		default:
			return nil, sdk.ErrUnknownRequest("query symbol : " + path[0])
		}
	}
}

func queryParameters2(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}
func queryMarket(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}
func queryMarketList(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}
func queryOrdersInMarket(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}
func queryOrder(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}
func queryUserOrderList(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}
func queryWaitCancelMarkets(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}
