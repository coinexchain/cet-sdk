package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math"
)

var _ sdk.Msg = MsgCreateTradingPair{}
var _ sdk.Msg = MsgCreateOrder{}
var _ sdk.Msg = MsgCancelOrder{}
var _ sdk.Msg = MsgAddLiquidity{}
var _ sdk.Msg = MsgRemoveLiquidity{}
var _ sdk.Msg = MsgCancelTradingPair{}

type MsgCreateTradingPair struct {
	Stock          string         `json:"stock"`
	Money          string         `json:"money"`
	Creator        sdk.AccAddress `json:"creator"`
	PricePrecision byte           `json:"price_precision"`
}

func (m MsgCreateTradingPair) Route() string {
	return ModuleName
}

func (m MsgCreateTradingPair) Type() string {
	return "create_pair"
}

func (m MsgCreateTradingPair) ValidateBasic() sdk.Error {
	if m.Creator.Empty() {
		return sdk.ErrInvalidAddress("missing creator address")
	}
	if len(m.Stock) == 0 || len(m.Money) == 0 {
		return ErrInvalidToken("token is empty")
	}
	return nil
}

func (m MsgCreateTradingPair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgCreateTradingPair) GetSigners() []sdk.AccAddress {
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

type MsgCreateOrder struct {
	Sender         sdk.AccAddress `json:"sender"`
	Identify       byte           `json:"identify"`
	TradingPair    string         `json:"trading_pair"`
	PricePrecision byte           `json:"price_precision"`
	Price          int64          `json:"price"`
	Quantity       int64          `json:"quantity"`
	Side           byte           `json:"side"`
}

func (m MsgCreateOrder) Route() string {
	panic("implement me")
}

func (m MsgCreateOrder) Type() string {
	panic("implement me")
}

func (m MsgCreateOrder) ValidateBasic() sdk.Error {
	panic("implement me")
}

func (m MsgCreateOrder) GetSignBytes() []byte {
	panic("implement me")
}

func (m MsgCreateOrder) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

func (m MsgCreateOrder) GetOrder() *Order {
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

type MsgCancelOrder struct {
	Sender  sdk.AccAddress `json:"sender"`
	OrderID string         `json:"order_id"`
}

func (m MsgCancelOrder) Route() string {
	panic("implement me")
}

func (m MsgCancelOrder) Type() string {
	panic("implement me")
}

func (m MsgCancelOrder) ValidateBasic() sdk.Error {
	panic("implement me")
}

func (m MsgCancelOrder) GetSignBytes() []byte {
	panic("implement me")
}

func (m MsgCancelOrder) GetSigners() []sdk.AccAddress {
	panic("implement me")
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
