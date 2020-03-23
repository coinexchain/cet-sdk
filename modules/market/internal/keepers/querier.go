package keepers

import (
	"fmt"
	"math"
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/market/internal/types"
)

const (
	QueryMarket            = "market-info"
	QueryMarkets           = "market-list"
	QueryOrdersInMarket    = "orders-in-market"
	QueryOrder             = "order-info"
	QueryUserOrders        = "user-order-list"
	QueryWaitCancelMarkets = "wait-cancel-markets"
	QueryParameters        = "parameters"
)

// creates a querier for asset REST endpoints
func NewQuerier(mk Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryParameters:
			return queryParameters(ctx, mk)
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

func queryParameters(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

type QueryMarketParam struct {
	TradingPair string
}

func NewQueryMarketParam(symbol string) QueryMarketParam {
	return QueryMarketParam{
		TradingPair: symbol,
	}
}

type QueryMarketInfo struct {
	Creator           sdk.AccAddress `json:"creator"`
	Stock             string         `json:"stock"`
	Money             string         `json:"money"`
	PricePrecision    string         `json:"price_precision"`
	LastExecutedPrice sdk.Dec        `json:"last_executed_price"`
	OrderPrecision    string         `json:"order_precision"`
}

func queryMarket(ctx sdk.Context, req abci.RequestQuery, mk Keeper) ([]byte, sdk.Error) {
	var param QueryMarketParam
	if err := mk.cdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse param: %s", err))
	}

	info, err := mk.GetMarketInfo(ctx, param.TradingPair)
	if err != nil {
		return nil, types.ErrInvalidMarket("Maybe the market have been deleted or not exist")
	}

	queryInfo := QueryMarketInfo{
		Creator:           mk.MarketOwner(ctx, info),
		Stock:             info.Stock,
		Money:             info.Money,
		PricePrecision:    strconv.Itoa(int(info.PricePrecision)),
		LastExecutedPrice: info.LastExecutedPrice,
		OrderPrecision:    strconv.Itoa(int(info.OrderPrecision)),
	}
	bz, err := codec.MarshalJSONIndent(mk.cdc, queryInfo)
	if err != nil {
		return nil, types.ErrFailedMarshal()
	}
	return bz, nil
}

type MarketInfoList []QueryMarketInfo

func queryMarketList(ctx sdk.Context, req abci.RequestQuery, mk Keeper) ([]byte, sdk.Error) {
	infos := mk.GetAllMarketInfos(ctx)
	mInfoList := make([]QueryMarketInfo, len(infos))

	for i, info := range infos {
		mInfoList[i] = QueryMarketInfo{
			Creator:           mk.MarketOwner(ctx, info),
			Stock:             info.Stock,
			Money:             info.Money,
			PricePrecision:    strconv.Itoa(int(info.PricePrecision)),
			LastExecutedPrice: info.LastExecutedPrice,
			OrderPrecision:    strconv.Itoa(int(info.OrderPrecision)),
		}
	}
	bz, err := codec.MarshalJSONIndent(mk.cdc, mInfoList)
	if err != nil {
		return nil, types.ErrFailedMarshal()
	}
	return bz, nil
}

type ResOrder struct {
	OrderID          string
	Sender           sdk.AccAddress `json:"sender"`
	Sequence         uint64         `json:"sequence"`
	Identify         byte           `json:"identify"`
	TradingPair      string         `json:"trading_pair"`
	OrderType        byte           `json:"order_type"`
	Price            sdk.Dec        `json:"price"`
	Quantity         int64          `json:"quantity"`
	Side             byte           `json:"side"`
	TimeInForce      int64          `json:"time_in_force"`
	Height           int64          `json:"height"`
	FrozenCommission int64          `json:"frozen_commission"` // DEX2
	ExistBlocks      int64          `json:"exist_blocks"`
	FrozenFeatureFee int64          `json:"frozen_feature_fee"`   // DEX2
	FrozenFee        int64          `json:"frozen_fee,omitempty"` // DEX2: -> frozen_commission

	// These fields will change when order was filled/canceled.
	LeftStock int64 `json:"left_stock"`
	Freeze    int64 `json:"freeze"`
	DealStock int64 `json:"deal_stock"`
	DealMoney int64 `json:"deal_money"`
}

func convertResOrderFromOrder(order *types.Order) *ResOrder {
	return &ResOrder{
		OrderID:          order.OrderID(),
		Sender:           order.Sender,
		Sequence:         order.Sequence,
		Identify:         order.Identify,
		TradingPair:      order.TradingPair,
		OrderType:        order.OrderType,
		Price:            order.Price,
		Quantity:         order.Quantity,
		Side:             order.Side,
		TimeInForce:      order.TimeInForce,
		Height:           order.Height,
		FrozenCommission: order.FrozenCommission,
		ExistBlocks:      order.ExistBlocks,
		FrozenFeatureFee: order.FrozenFeatureFee,
		FrozenFee:        order.FrozenFee,
		LeftStock:        order.LeftStock,
		Freeze:           order.Freeze,
		DealStock:        order.DealStock,
		DealMoney:        order.DealMoney,
	}
}

func queryOrdersInMarket(ctx sdk.Context, req abci.RequestQuery, mk Keeper) ([]byte, sdk.Error) {
	var param QueryMarketParam
	if err := mk.cdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse param: %s", err))
	}

	k := NewOrderKeeper(mk.marketKey, param.TradingPair, mk.cdc)
	orders := k.GetOlderThan(ctx, math.MaxInt64)
	rs := make([]*ResOrder, len(orders))
	for i, or := range orders {
		rs[i] = convertResOrderFromOrder(or)
	}
	bz, err := codec.MarshalJSONIndent(mk.cdc, rs)
	if err != nil {
		return nil, types.ErrFailedMarshal()
	}
	return bz, nil
}

type QueryOrderParam struct {
	OrderID string
}

func NewQueryOrderParam(orderID string) QueryOrderParam {
	return QueryOrderParam{
		OrderID: orderID,
	}
}

func queryOrder(ctx sdk.Context, req abci.RequestQuery, mk Keeper) ([]byte, sdk.Error) {
	var param QueryOrderParam
	if err := mk.cdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, types.ErrFailedParseParam()
	}

	okp := NewGlobalOrderKeeper(mk.marketKey, mk.cdc)
	order := okp.QueryOrder(ctx, param.OrderID)
	if order == nil {
		return nil, types.ErrOrderNotFound(param.OrderID)
	}
	bz, err := codec.MarshalJSONIndent(mk.cdc, *order)
	if err != nil {
		return nil, types.ErrFailedMarshal()
	}

	return bz, nil
}

type QueryUserOrderList struct {
	User string
}

func queryUserOrderList(ctx sdk.Context, req abci.RequestQuery, mk Keeper) ([]byte, sdk.Error) {
	var param QueryUserOrderList
	if err := mk.cdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, types.ErrFailedParseParam()
	}

	okp := NewGlobalOrderKeeper(mk.marketKey, mk.cdc)
	orders := okp.GetOrdersFromUser(ctx, param.User)
	if len(orders) == 0 {
		orders = append(orders, "")
	}
	bz, err := codec.MarshalJSONIndent(mk.cdc, orders)
	if err != nil {
		return nil, types.ErrFailedMarshal()
	}

	return bz, nil
}

type QueryCancelMarkets struct {
	Time int64
}

func queryWaitCancelMarkets(ctx sdk.Context, req abci.RequestQuery, mk Keeper) ([]byte, sdk.Error) {
	var param QueryCancelMarkets
	if err := mk.cdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, types.ErrFailedParseParam()
	}

	dlk := NewDelistKeeper(mk.marketKey)
	markets := dlk.GetDelistSymbolsBeforeTime(ctx, param.Time)
	bz, err := codec.MarshalJSONIndent(mk.cdc, markets)
	if err != nil {
		return nil, types.ErrFailedMarshal()
	}
	return bz, nil
}
