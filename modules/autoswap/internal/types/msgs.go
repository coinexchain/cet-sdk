package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreateLimitOrder{}
var _ sdk.Msg = MsgCreateMarketOrder{}
var _ sdk.Msg = MsgDeleteOrder{}
var _ sdk.Msg = MsgAddLiquidity{}
var _ sdk.Msg = MsgRemoveLiquidity{}

type MsgCreateLimitOrder struct {
	OrderBasic `json:"order_basic"`
	OrderID    int64    `json:"order_id"`
	Price      sdk.Dec  `json:"price"`
	PrevKey    [3]int64 `json:"prev_key"`
}

func (limit *MsgCreateLimitOrder) Route() string {
	return RouterKey
}

func (limit *MsgCreateLimitOrder) Type() string {
	return "create_limit_order"
}

func (limit *MsgCreateLimitOrder) ValidateBasic() sdk.Error {
	if len(strings.TrimSpace(limit.MarketSymbol)) == 0 || (!limit.IsOpenSwap && !limit.IsOpenOrderBook) {
		return ErrInvalidMarket(limit.MarketSymbol, limit.IsOpenSwap, limit.IsOpenOrderBook)
	}
	if limit.Sender.Empty() {
		return ErrInvalidSender(limit.Sender)
	}
	if limit.Price.IsZero() {
		return ErrInvalidPrice(limit.Price.String())
	}
	actualAmount := limit.OrderInfo().ActualAmount()
	if actualAmount.GT(MaxAmount) || actualAmount.IsZero() {
		return ErrInvalidAmount(actualAmount)
	}
	return nil
}

func (limit *MsgCreateLimitOrder) OrderInfo() *Order {
	return &Order{
		OrderBasic: limit.OrderBasic,
		Price:      limit.Price,
		PrevKey:    limit.PrevKey,
		OrderID:    limit.OrderID,
	}
}

func (limit *MsgCreateLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(limit))
}

func (limit *MsgCreateLimitOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{limit.Sender}
}

func (limit *MsgCreateLimitOrder) SetAccAddress(address sdk.AccAddress) {
	limit.Sender = address
}

type MsgCreateMarketOrder struct {
	OrderBasic      `json:"order_basic"`
	MinOutputAmount sdk.Int `json:"min_output_amount"`
}

func (mkOr MsgCreateMarketOrder) Route() string {
	return RouterKey
}

func (mkOr MsgCreateMarketOrder) Type() string {
	return "create_market_order"
}

func (mkOr MsgCreateMarketOrder) ValidateBasic() sdk.Error {
	if len(strings.TrimSpace(mkOr.MarketSymbol)) == 0 || (!mkOr.IsOpenSwap && !mkOr.IsOpenOrderBook) {
		return ErrInvalidMarket(mkOr.MarketSymbol, mkOr.IsOpenSwap, mkOr.IsOpenOrderBook)
	}
	if mkOr.Sender.Empty() {
		return ErrInvalidSender(mkOr.Sender)
	}
	if mkOr.Amount.IsZero() || mkOr.Amount.IsNegative() {
		return ErrInvalidAmount(mkOr.Amount)
	}
	if mkOr.MinOutputAmount.IsNegative() {
		return ErrInvalidOutputAmount(mkOr.MinOutputAmount)
	}
	return nil
}

func (mkOr MsgCreateMarketOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(mkOr))
}

func (mkOr MsgCreateMarketOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{mkOr.Sender}
}

func (mkOr *MsgCreateMarketOrder) SetAccAddress(address sdk.AccAddress) {
	mkOr.Sender = address
}

func (mkOr *MsgCreateMarketOrder) OrderInfo() *Order {
	return &Order{
		OrderBasic:      mkOr.OrderBasic,
		MinOutputAmount: mkOr.MinOutputAmount,
	}
}

type MsgDeleteOrder struct {
	MarketSymbol    string         `json:"market_symbol"`
	IsOpenSwap      bool           `json:"is_open_swap"`
	IsOpenOrderBook bool           `json:"is_open_order_book"`
	Sender          sdk.AccAddress `json:"sender"`
	IsBuy           bool           `json:"is_buy"`
	PrevKey         [3]int64       `json:"prev_key"`
	OrderID         int64          `json:"order_id"`
}

func (m MsgDeleteOrder) Route() string {
	return RouterKey
}

func (m MsgDeleteOrder) Type() string {
	return "delete_order"
}

func (m MsgDeleteOrder) ValidateBasic() sdk.Error {
	if len(strings.TrimSpace(m.MarketSymbol)) == 0 || (!m.IsOpenSwap && !m.IsOpenOrderBook) {
		return ErrInvalidMarket(m.MarketSymbol, m.IsOpenSwap, m.IsOpenOrderBook)
	}
	if m.OrderID <= 0 {
		return ErrInvalidOrderID(m.OrderID)
	}
	if m.Sender.Empty() {
		return ErrInvalidSender(m.Sender)
	}
	return nil
}

func (m MsgDeleteOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgDeleteOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

func (m *MsgDeleteOrder) SetAccAddress(address sdk.AccAddress) {
	m.Sender = address
}

func (m MsgDeleteOrder) OrderInfo() *Order {
	return &Order{
		OrderBasic: OrderBasic{
			Sender:          m.Sender,
			MarketSymbol:    m.MarketSymbol,
			IsOpenSwap:      m.IsOpenSwap,
			IsOpenOrderBook: m.IsOpenOrderBook,
			IsBuy:           m.IsBuy,
		},
		OrderID: m.OrderID,
		PrevKey: m.PrevKey,
	}
}

type MsgAddLiquidity struct {
	Owner      sdk.AccAddress `json:"owner"`
	Stock      string         `json:"stock"`
	Money      string         `json:"money"`
	StockIn    sdk.Int        `json:"stock_in"`
	MoneyIn    sdk.Int        `json:"money_in"`
	IsOpenSwap bool           `json:"is_open_swap"`
	To         sdk.AccAddress `json:"to"`
}

func (m MsgAddLiquidity) Route() string {
	return RouterKey
}

func (m MsgAddLiquidity) Type() string {
	return "add_liquidity"
}

func (m MsgAddLiquidity) ValidateBasic() sdk.Error {
	if m.Owner.Empty() {
		return sdk.ErrInvalidAddress("missing owner address")
	}
	if len(m.Stock) == 0 || len(m.Money) == 0 {
		//todo:
		return nil
	}
	if m.StockIn.IsZero() && m.MoneyIn.IsPositive() || m.MoneyIn.IsZero() && m.StockIn.IsPositive() {
		return nil
	}
	//if To is nil, Owner => To
	return nil
}

func (m MsgAddLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

func (m *MsgAddLiquidity) SetAccAddress(address sdk.AccAddress) {
	m.Owner = address
}

type MsgRemoveLiquidity struct {
	Sender         sdk.AccAddress `json:"sender"`
	Stock          string         `json:"stock"`
	Money          string         `json:"money"`
	AmmOpen        bool           `json:"amm_open"`
	PoolOpen       bool           `json:"pool_open"`
	Amount         sdk.Int        `json:"amount"`
	To             sdk.AccAddress `json:"to"`
	AmountStockMin sdk.Int        `json:"amount_stock_min"`
	AmountMoneyMin sdk.Int        `json:"amount_money_min"`
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
		//todo:
		return nil
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
