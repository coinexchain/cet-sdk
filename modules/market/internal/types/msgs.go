package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/asset"
	dex "github.com/coinexchain/cet-sdk/types"
)

// RouterKey is the name of the market module
const (
	// msg keys for Kafka
	CreateMarketInfoKey = "create_market_info"
	CancelMarketInfoKey = "cancel_market_info"
	CreateOrderInfoKey  = "create_order_info"
	FillOrderInfoKey    = "fill_order_info"
	CancelOrderInfoKey  = "del_order_info"
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

// /////////////////////////////////////////////////////////
// MsgCreateTradingPair

var _ sdk.Msg = MsgCreateTradingPair{}

type MsgCreateTradingPair struct {
	Stock          string         `json:"stock"`
	Money          string         `json:"money"`
	Creator        sdk.AccAddress `json:"creator"`
	PricePrecision byte           `json:"price_precision"`
	OrderPrecision byte           `json:"order_precision"` // not used
}

func NewMsgCreateTradingPair(stock, money string, creator sdk.AccAddress, pricePrecision byte, orderPrecision byte) MsgCreateTradingPair {
	return MsgCreateTradingPair{
		Stock:          stock,
		Money:          money,
		Creator:        creator,
		PricePrecision: pricePrecision,
		OrderPrecision: orderPrecision,
	}
}

func (msg MsgCreateTradingPair) GetSymbol() string {
	return dex.GetSymbol(msg.Stock, msg.Money)
}

// --------------------------------------------------------
// sdk.Msg Implementation

func (msg *MsgCreateTradingPair) SetAccAddress(address sdk.AccAddress) {
	msg.Creator = address
}

func (msg MsgCreateTradingPair) Route() string { return RouterKey }

func (msg MsgCreateTradingPair) Type() string { return "create_market_info" }

func (msg MsgCreateTradingPair) ValidateBasic() sdk.Error {
	if err := sdk.VerifyAddressFormat(msg.Creator); err != nil {
		return ErrInvalidAddress()
	}
	if !IsValidTradingPair([]string{msg.Stock, msg.Money}) {
		return ErrInvalidSymbol()
	}
	if p := msg.PricePrecision; p > MaxTokenPricePrecision {
		return ErrInvalidPricePrecision(p)
	}
	if msg.Money == msg.Stock {
		return ErrStockAndMoneyAreSame()
	}
	return nil
}

func (msg MsgCreateTradingPair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgCreateTradingPair) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Creator}
}

// /////////////////////////////////////////////////////////
// MsgCreateOrder

var _ sdk.Msg = MsgCreateOrder{}

type MsgCreateOrder struct {
	Sender         sdk.AccAddress `json:"sender"`
	Identify       byte           `json:"identify"`
	TradingPair    string         `json:"trading_pair"`
	OrderType      byte           `json:"order_type"`
	PricePrecision byte           `json:"price_precision"`
	Price          int64          `json:"price"`
	Quantity       int64          `json:"quantity"`
	Side           byte           `json:"side"`
	TimeInForce    int64          `json:"time_in_force"`
	ExistBlocks    int64          `json:"exist_blocks"`
}

func (msg *MsgCreateOrder) SetAccAddress(address sdk.AccAddress) {
	msg.Sender = address
}

func (msg MsgCreateOrder) Route() string { return RouterKey }

func (msg MsgCreateOrder) Type() string { return "create_order" }

func (msg MsgCreateOrder) ValidateBasic() sdk.Error {
	if err := sdk.VerifyAddressFormat(msg.Sender); err != nil {
		return ErrInvalidAddress()
	}
	if !IsValidTradingPair(strings.Split(msg.TradingPair, SymbolSeparator)) {
		return ErrInvalidSymbol()
	}
	if msg.OrderType != LimitOrder {
		return ErrInvalidOrderType()
	}
	if p := msg.PricePrecision; p > MaxTokenPricePrecision {
		return ErrInvalidPricePrecision(p)
	}
	if msg.Price <= 0 {
		return ErrInvalidPrice(msg.Price)
	}
	if msg.Quantity <= 0 {
		return ErrOrderAmountTooSmall(fmt.Sprintf("%d", msg.Quantity))
	}
	if msg.Side != BUY && msg.Side != SELL {
		return ErrInvalidTradeSide()
	}
	if msg.TimeInForce != GTE && msg.TimeInForce != IOC {
		return ErrInvalidTimeInForce(msg.TimeInForce)
	}
	if msg.ExistBlocks < 0 {
		return ErrInvalidExistBlocks(msg.ExistBlocks)
	}

	return nil
}

func (msg MsgCreateOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgCreateOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
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

func (msg *MsgCancelOrder) SetAccAddress(address sdk.AccAddress) {
	msg.Sender = address
}

func (msg MsgCancelOrder) Route() string {
	return StoreKey
}

func (msg MsgCancelOrder) Type() string {
	return "cancel_order"
}

func ValidateOrderID(id string) sdk.Error {
	contents := strings.Split(id, OrderIDSeparator)
	if len(contents) != OrderIDPartsNum {
		return ErrInvalidOrderID()
	}
	if seqWithIdenti, err := strconv.Atoi(contents[1]); err != nil || seqWithIdenti < 0 {
		return ErrInvalidOrderID()
	}
	if _, err := sdk.AccAddressFromBech32(contents[0]); err != nil {
		return ErrInvalidAddress()
	}
	return nil
}

func (msg MsgCancelOrder) ValidateBasic() sdk.Error {
	if err := sdk.VerifyAddressFormat(msg.Sender); err != nil {
		return ErrInvalidAddress()
	}
	if err := ValidateOrderID(msg.OrderID); err != nil {
		return err
	}
	return nil
}

func (msg MsgCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgCancelOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// /////////////////////////////////////////////////////////
// MsgCancelTradingPair

type MsgCancelTradingPair struct {
	Sender        sdk.AccAddress `json:"sender"`
	TradingPair   string         `json:"trading_pair"`
	EffectiveTime int64          `json:"effective_time"`
}

func (msg *MsgCancelTradingPair) SetAccAddress(address sdk.AccAddress) {
	msg.Sender = address
}

func (msg MsgCancelTradingPair) Route() string {
	return RouterKey
}

func (msg MsgCancelTradingPair) Type() string {
	return "cancel_market"
}

func IsValidTradingPair(tokens []string) bool {
	if len(tokens) != 2 {
		return false
	}
	if err := asset.ValidateTokenSymbol(tokens[0]); err != nil {
		return false
	}
	if err := asset.ValidateTokenSymbol(tokens[1]); err != nil {
		return false
	}
	return true
}

func (msg MsgCancelTradingPair) ValidateBasic() sdk.Error {
	if err := sdk.VerifyAddressFormat(msg.Sender); err != nil {
		return ErrInvalidAddress()
	}
	if !IsValidTradingPair(strings.Split(msg.TradingPair, SymbolSeparator)) {
		return ErrInvalidSymbol()
	}
	if msg.EffectiveTime < 0 {
		return ErrInvalidCancelTime()
	}

	return nil
}

func (msg MsgCancelTradingPair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgCancelTradingPair) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// -------------------------------------------------
// MsgModifyPricePrecision

type MsgModifyPricePrecision struct {
	Sender         sdk.AccAddress `json:"sender"`
	TradingPair    string         `json:"trading_pair"`
	PricePrecision byte           `json:"price_precision"`
}

func (msg *MsgModifyPricePrecision) SetAccAddress(address sdk.AccAddress) {
	msg.Sender = address
}

func (msg MsgModifyPricePrecision) Route() string {
	return RouterKey
}

func (msg MsgModifyPricePrecision) Type() string {
	return "modify_trading_pair_price_precision"
}

func (msg MsgModifyPricePrecision) ValidateBasic() sdk.Error {
	if err := sdk.VerifyAddressFormat(msg.Sender); err != nil {
		return ErrInvalidAddress()
	}
	if !IsValidTradingPair(strings.Split(msg.TradingPair, SymbolSeparator)) {
		return ErrInvalidSymbol()
	}
	if p := msg.PricePrecision; p < MinTokenPricePrecision || p > MaxTokenPricePrecision {
		return ErrInvalidPricePrecision(p)
	}
	return nil
}

func (msg MsgModifyPricePrecision) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgModifyPricePrecision) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
