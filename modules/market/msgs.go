package market

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/dex/modules/asset"
	"github.com/coinexchain/dex/modules/market/match"
)

// RouterKey is the name of the market module
const (
	RouterKey = "market"
	StoreKey  = RouterKey
	Topic     = RouterKey

	// msg keys for Kafka
	CreateMarketInfoKey = "create_market_info"
	CancelMarketInfoKey = "cancel_market_info"

	CreateOrderInfoKey = "create_order_info"
	FillOrderInfoKey   = "fill_order_info"
	CancelOrderInfoKey = "del_order_info"
)

// cancel order of reasons
const (
	CancelOrderByManual        = "Manually cancel the order"
	CancelOrderByAllFilled     = "The order was fully filled"
	CancelOrderByGteTimeOut    = "GTE order timeout"
	CancelOrderByIocType       = "IOC order cancel "
	CancelOrderByNoEnoughMoney = "Insufficient freeze money"
	CancelOrderByNotKnow       = "Don't know"
)

var (
	msgCdc = codec.New()
)

func init() {
	RegisterCodec(msgCdc)
}

// /////////////////////////////////////////////////////////
// MsgCreateMarketInfo

var _ sdk.Msg = MsgCreateMarketInfo{}

type MsgCreateMarketInfo struct {
	Stock          string         `json:"stock"`
	Money          string         `json:"money"`
	Creator        sdk.AccAddress `json:"creator"`
	PricePrecision byte           `json:"price_precision"`
}

func NewMsgCreateMarketInfo(stock, money string, crater sdk.AccAddress, pricePrecision byte) MsgCreateMarketInfo {
	return MsgCreateMarketInfo{
		Stock:          stock,
		Money:          money,
		Creator:        crater,
		PricePrecision: pricePrecision,
	}
}

// --------------------------------------------------------
// sdk.Msg Implementation

func (msg MsgCreateMarketInfo) Route() string { return RouterKey }

func (msg MsgCreateMarketInfo) Type() string { return "create_market_info" }

func (msg MsgCreateMarketInfo) ValidateBasic() sdk.Error {
	if len(msg.Creator) == 0 {
		return sdk.ErrInvalidAddress("missing creator address")
	}
	if len(msg.Stock) == 0 || len(msg.Money) == 0 {
		return sdk.ErrInvalidAddress("missing stock or money identifier")
	}
	if msg.PricePrecision > sdk.Precision {
		return sdk.ErrInvalidAddress("price precision value out of range [0, 18]")
	}
	return nil
}

func (msg MsgCreateMarketInfo) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgCreateMarketInfo) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{[]byte(msg.Creator)}
}

// /////////////////////////////////////////////////////////
// MsgCreateOrder

var _ sdk.Msg = MsgCreateOrder{}

type MsgCreateOrder struct {
	Sender         sdk.AccAddress `json:"sender"`
	Sequence       uint64         `json:"sequence"`
	Symbol         string         `json:"symbol"`
	OrderType      byte           `json:"order_type"`
	PricePrecision byte           `json:"price_precision"`
	Price          int64          `json:"price"`
	Quantity       int64          `json:"quantity"`
	Side           byte           `json:"side"`
	TimeInForce    int            `json:"time_in_force"`
}

func (msg MsgCreateOrder) Route() string { return RouterKey }

func (msg MsgCreateOrder) Type() string { return "create_order" }

func (msg MsgCreateOrder) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrInvalidAddress("missing creator address")
	}
	if len(msg.Symbol) == 0 {
		return sdk.ErrInvalidAddress("missing GTE order symbol identifier")
	}
	if msg.PricePrecision < MinTokenPricePrecision ||
		msg.PricePrecision > MaxTokenPricePrecision {
		return sdk.ErrInvalidAddress(fmt.Sprintf("price precision value out of range [8, 18]. actual : %d", msg.PricePrecision))
	}

	if msg.Side != match.BUY && msg.Side != match.SELL {
		return ErrInvalidTradeSide()
	}

	if msg.OrderType != LimitOrder {
		return ErrInvalidOrderType()
	}

	if len(strings.Split(msg.Symbol, SymbolSeparator)) != 2 {
		return ErrInvalidSymbol()
	}

	if msg.Price <= 0 || msg.Price > asset.MaxTokenAmount {
		return ErrInvalidPrice()
	}

	return nil
}

func (msg MsgCreateOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgCreateOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{[]byte(msg.Sender)}
}

func (msg MsgCreateOrder) IsGTEOrder() bool {
	return msg.TimeInForce == GTE
}

// /////////////////////////////////////////////////////////
// MsgCancelOrder

type MsgCancelOrder struct {
	Sender  sdk.AccAddress `json:"sender"`
	OrderID string         `json:"order_id"`
}

func (msg MsgCancelOrder) Route() string {
	return StoreKey
}

func (msg MsgCancelOrder) Type() string {
	return "cancel_order"
}

func (msg MsgCancelOrder) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return ErrInvalidAddress()
	}

	if len(strings.Split(msg.OrderID, "-")) != 2 {
		return ErrInvalidOrderID()
	}

	return nil
}

func (msg MsgCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgCancelOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// /////////////////////////////////////////////////////////
// MsgCancelMarket

type MsgCancelMarket struct {
	Sender          sdk.AccAddress `json:"sender"`
	Symbol          string         `json:"symbol"`
	EffectiveHeight int64          `json:"effective_height"`
}

func (msg MsgCancelMarket) Route() string {
	return StoreKey
}

func (msg MsgCancelMarket) Type() string {
	return "cancel_market"
}

func (msg MsgCancelMarket) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return ErrInvalidAddress()
	}

	if len(strings.Split(msg.Symbol, SymbolSeparator)) != 2 {
		return ErrInvalidSymbol()
	}

	if msg.EffectiveHeight < 0 {
		return sdk.NewError(CodeSpaceMarket, CodeInvalidHeight, "Invalid height")
	}

	return nil
}

func (msg MsgCancelMarket) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg MsgCancelMarket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// --------------------------------------------------------------------
// msg queue infos for kafka

type CreateMarketInfo struct {
	Stock          string `json:"stock"`
	Money          string `json:"money"`
	PricePrecision byte   `json:"price_precision"`

	// create market info
	Creator      string `json:"creator"`
	CreateHeight int64  `json:"create_height"`
}

type CancelMarketInfo struct {
	Stock string `json:"stock"`
	Money string `json:"money"`

	// del market info
	Deleter string `json:"deleter"`
	DelTime int64  `json:"del_time"`
}

type CreateOrderInfo struct {
	OrderID     string `json:"order_id"`
	Sender      string `json:"sender"`
	Symbol      string `json:"symbol"`
	OrderType   byte   `json:"order_type"`
	Price       string `json:"price"`
	Quantity    int64  `json:"quantity"`
	Side        byte   `json:"side"`
	TimeInForce int    `json:"time_in_force"`
	Height      int64  `json:"height"`
	FrozenFee   int64  `json:"frozen_fee"`
	Freeze      int64  `json:"freeze"`
}

type FillOrderInfo struct {
	OrderID string `json:"order_id"`
	Height  int64  `json:"height"`

	// These fields will change when order was filled/canceled.
	LeftStock int64 `json:"left_stock"`
	Freeze    int64 `json:"freeze"`
	DealStock int64 `json:"deal_stock"`
	DealMoney int64 `json:"deal_money"`
	CurrStock int64 `json:"curr_stock"`
}

type CancelOrderInfo struct {
	OrderID string `json:"order_id"`

	// Del infos
	DelReason string `json:"del_reason"`
	DelHeight int64  `json:"del_height"`

	// Fields of amount
	UseFee       int64 `json:"use_fee"`
	LeftStock    int64 `json:"left_stock"`
	RemainAmount int64 `json:"remain_amount"`
	DealStock    int64 `json:"deal_stock"`
	DealMoney    int64 `json:"deal_money"`
}
