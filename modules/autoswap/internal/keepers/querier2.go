package keepers

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cet-sdk/modules/market"
	dex "github.com/coinexchain/cet-sdk/types"
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
	var param market.QueryMarketParam
	if err := k.cdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse param: %s", err))
	}

	info := k.GetPoolInfo(ctx, param.TradingPair)
	if info == nil {
		return nil, types.ErrInvalidMarket("Maybe the market have been deleted or not exist")
	}

	queryInfo := toMarketQueryInfo(info)
	bz, err := codec.MarshalJSONIndent(k.cdc, queryInfo)
	if err != nil {
		return nil, types.ErrMarshalFailed()
	}
	return bz, nil
}
func queryMarketList(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	infos := k.GetPoolInfos(ctx)
	mInfoList := make([]market.QueryMarketInfo, len(infos))

	for i, info := range infos {
		mInfoList[i] = toMarketQueryInfo(&info)
	}
	bz, err := codec.MarshalJSONIndent(k.cdc, mInfoList)
	if err != nil {
		return nil, types.ErrMarshalFailed()
	}
	return bz, nil
}
func queryOrdersInMarket(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var param market.QueryMarketParam
	if err := k.cdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse param: %s", err))
	}

	//orders := k.GetOrder(ctx, math.MaxInt64)
	var orders []*types.Order // TODO
	rs := make([]*market.ResOrder, len(orders))
	for i, order := range orders {
		rs[i] = toResOrder(order)
	}
	bz, err := codec.MarshalJSONIndent(k.cdc, rs)
	if err != nil {
		return nil, types.ErrMarshalFailed()
	}
	return bz, nil
}
func queryOrder(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var param market.QueryOrderParam
	if err := k.cdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, types.ErrMarshalFailed()
	}

	order := k.GetOrder(ctx, param.OrderID)
	if order == nil {
		return nil, types.ErrNotFoundOrder(param.OrderID)
	}
	bz, err := codec.MarshalJSONIndent(k.cdc, *order)
	if err != nil {
		return nil, types.ErrMarshalFailed()
	}

	return bz, nil
}
func queryUserOrderList(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}
func queryWaitCancelMarkets(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	panic("TODO")
}

func toMarketQueryInfo(info *PoolInfo) market.QueryMarketInfo {
	queryInfo := market.QueryMarketInfo{
		Creator:           info.Owner,
		PricePrecision:    strconv.Itoa(int(info.PricePrecision)),
		LastExecutedPrice: sdk.NewDec(0), // TODO
		OrderPrecision:    "0",           // not used
	}
	queryInfo.Stock, queryInfo.Money = dex.SplitSymbol(info.Symbol)
	return queryInfo
}

// TODO
func toResOrder(order *types.Order) *market.ResOrder {
	return &market.ResOrder{
		OrderID:     order.GetOrderID(),
		Sender:      order.Sender,
		Sequence:    0, // TODO
		Identify:    order.Identify,
		TradingPair: order.TradingPair,
		OrderType:   market.LimitOrder,
		Price:       order.Price,
		Quantity:    order.Quantity,
		Side:        order.GetSide(),
		Height:      order.Height,
		LeftStock:   order.LeftStock,
		Freeze:      order.Freeze,
		DealStock:   order.DealStock,
		DealMoney:   order.DealMoney,
		// not used fields
		//TimeInForce:      order.TimeInForce,
		//FrozenCommission: order.FrozenCommission,
		//ExistBlocks:      order.ExistBlocks,
		//FrozenFeatureFee: order.FrozenFeatureFee,
		//FrozenFee:        order.FrozenFee,
	}
}
