package market

import (
	"github.com/coinexchain/cet-sdk/modules/market/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/market/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

const (
	IntegrationNetSubString = types.IntegrationNetSubString
	OrderIDPartsNum         = types.OrderIDPartsNum
	SymbolSeparator         = types.SymbolSeparator
	LimitOrder              = types.LimitOrder
	GTE                     = types.GTE
	BID                     = types.BID
	ASK                     = types.ASK
	BUY                     = types.BUY
	SELL                    = types.SELL
	DecByteCount            = types.DecByteCount
	QueryMarket             = keepers.QueryMarket
	QueryMarkets            = keepers.QueryMarkets
	QueryOrdersInMarket     = keepers.QueryOrdersInMarket
	QueryOrder              = keepers.QueryOrder
	QueryUserOrders         = keepers.QueryUserOrders
	QueryWaitCancelMarkets  = keepers.QueryWaitCancelMarkets
	QueryParameters         = keepers.QueryParameters
)

var (
	NewBaseKeeper       = keepers.NewKeeper
	DefaultParams       = types.DefaultParams
	DecToBigEndianBytes = types.DecToBigEndianBytes
	ValidateOrderID     = types.ValidateOrderID
	IsValidTradingPair  = types.IsValidTradingPair
	ModuleCdc           = types.ModuleCdc
	GetSymbol           = dex.GetSymbol
	SplitSymbol         = dex.SplitSymbol
	AssemblyOrderID     = types.AssemblyOrderID
	RegisterCodec       = types.RegisterCodec
)

type (
	Keeper                  = keepers.Keeper
	Order                   = types.Order
	MarketInfo              = types.MarketInfo
	Params                  = types.Params
	MsgCreateOrder          = types.MsgCreateOrder
	MsgCreateTradingPair    = types.MsgCreateTradingPair
	MsgCancelOrder          = types.MsgCancelOrder
	MsgCancelTradingPair    = types.MsgCancelTradingPair
	MsgModifyPricePrecision = types.MsgModifyPricePrecision
	CreateOrderInfo         = types.CreateOrderInfo
	FillOrderInfo           = types.FillOrderInfo
	CancelOrderInfo         = types.CancelOrderInfo
	QueryMarketParam        = keepers.QueryMarketParam
	QueryOrderParam         = keepers.QueryOrderParam
	QueryMarketInfo         = keepers.QueryMarketInfo
	ResOrder                = keepers.ResOrder
)
