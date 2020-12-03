package types

import (
	"math"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = MsgAutoSwapCreateTradingPair{}
var _ sdk.Msg = MsgAutoSwapCreateOrder{}
var _ sdk.Msg = MsgAutoSwapCancelOrder{}
var _ sdk.Msg = MsgAddLiquidity{}
var _ sdk.Msg = MsgRemoveLiquidity{}
var _ sdk.Msg = MsgCancelTradingPair{}

type MsgAutoSwapCreateTradingPair struct {
	Stock          string         `json:"stock"`
	Money          string         `json:"money"`
	Creator        sdk.AccAddress `json:"creator"`
	PricePrecision byte           `json:"price_precision"`
}

func (m MsgAutoSwapCreateTradingPair) Route() string {
	return ModuleName
}

func (m MsgAutoSwapCreateTradingPair) Type() string {
	return "create_pair"
}

func (m MsgAutoSwapCreateTradingPair) ValidateBasic() sdk.Error {
	if m.Creator.Empty() {
		return sdk.ErrInvalidAddress("missing creator address")
	}
	if len(m.Stock) == 0 || len(m.Money) == 0 {
		return ErrInvalidToken("token is empty")
	}
	return nil
}

func (m MsgAutoSwapCreateTradingPair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgAutoSwapCreateTradingPair) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Creator}
}

type MsgCancelTradingPair struct {
	Sender        sdk.AccAddress `json:"sender"`
	TradingPair   string         `json:"trading_pair"`
	EffectiveTime int64          `json:"effective_time"`
}

func (m MsgCancelTradingPair) Route() string {
	return ModuleName
}

func (m MsgCancelTradingPair) Type() string {
	return "cancel_pair"
}

func (m MsgCancelTradingPair) ValidateBasic() sdk.Error {
	if m.Sender.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if len(m.TradingPair) == 0 {
		return ErrInvalidPairSymbol("pair symbol is invalid")
	}
	if m.EffectiveTime < 0 {
		return ErrInvalidEffectiveTime()
	}
	return nil
}

func (m MsgCancelTradingPair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgCancelTradingPair) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

type MsgAutoSwapCreateOrder struct {
	Sender         sdk.AccAddress `json:"sender"`
	Identify       byte           `json:"identify"`
	TradingPair    string         `json:"trading_pair"`
	PricePrecision byte           `json:"price_precision"`
	Price          int64          `json:"price"`
	Quantity       int64          `json:"quantity"`
	Side           byte           `json:"side"`
}

func (m MsgAutoSwapCreateOrder) Route() string {
	return ModuleName
}

func (m MsgAutoSwapCreateOrder) Type() string {
	return "create_order"
}

func (m MsgAutoSwapCreateOrder) ValidateBasic() sdk.Error {
	if m.Sender.Empty() {
		return ErrInvalidOrderSender(m.Sender)
	}
	if len(strings.Split(m.TradingPair, "/")) != 2 {
		return ErrInvalidMarket(m.TradingPair)
	}
	if m.Price <= 0 {
		return ErrInvalidPrice(m.Price)
	}
	actualAmount := m.GetOrder().ActualAmount()
	if actualAmount.GT(sdk.NewInt(math.MaxInt64)) || actualAmount.LTE(sdk.NewInt(0)) {
		return ErrInvalidOrderAmount(actualAmount)
	}
	if m.Side != BID && m.Side != ASK {
		return ErrInvalidOrderSide(m.Side)
	}
	return nil
}

func (m MsgAutoSwapCreateOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgAutoSwapCreateOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

func (m MsgAutoSwapCreateOrder) GetOrder() *Order {
	or := &Order{
		TradingPair:    m.TradingPair,
		Identify:       m.Identify,
		Sender:         m.Sender,
		Price:          sdk.NewDec(m.Price).Quo(sdk.NewDec(int64(math.Pow10(int(m.PricePrecision))))),
		Quantity:       m.Quantity,
		PricePrecision: m.PricePrecision,
		IsBuy:          m.Side == BID,
		LeftStock:      m.Quantity,
	}
	or.Freeze = or.ActualAmount().Int64()
	return or
}

type MsgAutoSwapCancelOrder struct {
	Sender  sdk.AccAddress `json:"sender"`
	OrderID string         `json:"order_id"`
}

func (m MsgAutoSwapCancelOrder) Route() string {
	return ModuleName
}

func (m MsgAutoSwapCancelOrder) Type() string {
	return "cancel_order"
}

func (m MsgAutoSwapCancelOrder) ValidateBasic() sdk.Error {
	if m.Sender.Empty() {
		return ErrInvalidOrderSender(m.Sender)
	}
	if len(m.OrderID) == 0 {
		return ErrInvalidOrderID(m.OrderID)
	}
	return nil
}

func (m MsgAutoSwapCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgAutoSwapCancelOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

type MsgAddLiquidity struct {
	Sender  sdk.AccAddress `json:"sender"`
	Stock   string         `json:"stock"`
	Money   string         `json:"money"`
	StockIn sdk.Int        `json:"stock_in"`
	MoneyIn sdk.Int        `json:"money_in"`
	To      sdk.AccAddress `json:"to"`
}

func (m MsgAddLiquidity) Route() string {
	return RouterKey
}

func (m MsgAddLiquidity) Type() string {
	return "add_liquidity"
}

func (m MsgAddLiquidity) ValidateBasic() sdk.Error {
	if m.Sender.Empty() {
		return sdk.ErrInvalidAddress("missing owner address")
	}
	if len(m.Stock) == 0 || len(m.Money) == 0 {
		return ErrInvalidToken("token is empty")
	}
	if !m.StockIn.IsPositive() {
		return ErrInvalidAmount(m.StockIn)
	}
	if !m.MoneyIn.IsPositive() {
		return ErrInvalidAmount(m.MoneyIn)
	}
	//if To is nil, Sender => To
	return nil
}

func (m MsgAddLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

func (m *MsgAddLiquidity) SetAccAddress(address sdk.AccAddress) {
	m.Sender = address
}

type MsgRemoveLiquidity struct {
	Sender sdk.AccAddress `json:"sender"`
	Stock  string         `json:"stock"`
	Money  string         `json:"money"`
	Amount sdk.Int        `json:"amount"`
	To     sdk.AccAddress `json:"to"`
}

func (m MsgRemoveLiquidity) Route() string {
	return RouterKey
}

func (m MsgRemoveLiquidity) Type() string {
	return "remove_liquidity"
}

func (m MsgRemoveLiquidity) ValidateBasic() sdk.Error {
	if m.Sender.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if len(m.Stock) == 0 || len(m.Money) == 0 {
		return ErrInvalidToken("token is empty")
	}
	if !m.Amount.IsPositive() {
		return ErrInvalidAmount(m.Amount)
	}
	//if To is nil, sender => To
	return nil
}

func (m MsgRemoveLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgRemoveLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

func (m *MsgRemoveLiquidity) SetAccAddress(address sdk.AccAddress) {
	m.Sender = address
}
