package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = MsgCreateLimitOrder{}
var _ sdk.Msg = MsgSwapTokens{}
var _ sdk.Msg = MsgDeleteOrder{}
var _ sdk.Msg = MsgAddLiquidity{}
var _ sdk.Msg = MsgRemoveLiquidity{}

type MsgCreateLimitOrder struct {
	OrderBasic `json:"order_basic"`
	OrderID    int64    `json:"order_id"`
	Price      sdk.Dec  `json:"price"`
	PrevKey    [3]int64 `json:"prev_key"`
}

func (limit MsgCreateLimitOrder) Route() string {
	return RouterKey
}

func (limit MsgCreateLimitOrder) Type() string {
	return "create_limit_order"
}

func (limit MsgCreateLimitOrder) ValidateBasic() sdk.Error {
	if len(strings.TrimSpace(limit.MarketSymbol)) == 0 {
		return ErrInvalidMarket(limit.MarketSymbol)
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

func (limit MsgCreateLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(limit))
}

func (limit MsgCreateLimitOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{limit.Sender}
}

func (limit *MsgCreateLimitOrder) SetAccAddress(address sdk.AccAddress) {
	limit.Sender = address
}

type MarketInfo struct {
	MarketSymbol    string // stock/money
}

type MsgSwapTokens struct {
	Pairs  []MarketInfo
	Sender sdk.AccAddress `json:"sender"`
	IsBuy  bool           `json:"is_buy"`

	// if the order is market_order, the amount is the actual input amount with special token(
	// ie: sell order, amount = stockTokenAmount, buy order = moneyTokenAmount)
	// if the order is limit_order, the amount is the stock amount and orderActualAmount will be calculated
	// (ie: buyActualAmount = price * amount, sellActualAmount = amount)
	Amount          sdk.Int `json:"amount"`
	MinOutputAmount sdk.Int `json:"min_output_amount"`
}

func (mkOr MsgSwapTokens) Route() string {
	return RouterKey
}

func (mkOr MsgSwapTokens) Type() string {
	return "create_market_order"
}

func (mkOr MsgSwapTokens) ValidateBasic() sdk.Error {
	if !isValidSwapChain(mkOr.Pairs) {
		return ErrInvalidSwap(mkOr.Pairs)
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

func isValidSwapChain(pairs []MarketInfo) bool {
	index := 0
	tokenLists := make([][]string, 0, len(pairs))
	if len(pairs) == 0 {
		return false
	}
	for _, v := range pairs {
		tokenLists = append(tokenLists, strings.Split(v.MarketSymbol, "/"))
		index = len(tokenLists) - 1
		if len(tokenLists[index]) != 2 || (tokenLists[index][0] == tokenLists[index][1]) {
			return false
		}
	}
	// swap pairs should be : a/b, b/c, c/d ....
	for i := 0; i < len(tokenLists)-1; i++ {
		if tokenLists[i][1] != tokenLists[i+1][0] {
			return false
		}
	}
	return true
}

func (mkOr MsgSwapTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(mkOr))
}

func (mkOr MsgSwapTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{mkOr.Sender}
}

func (mkOr *MsgSwapTokens) SetAccAddress(address sdk.AccAddress) {
	mkOr.Sender = address
}

func (mkOr *MsgSwapTokens) GetOrderInfos() []*Order {
	orders := make([]*Order, 0, len(mkOr.Pairs))
	for _, v := range mkOr.Pairs {
		orders = append(orders, &Order{
			OrderBasic: OrderBasic{
				Sender:          mkOr.Sender,
				MarketSymbol:    v.MarketSymbol,
				Amount:          mkOr.Amount,
			},
			MinOutputAmount: sdk.ZeroInt(),
		})
	}
	orders[len(orders)-1].MinOutputAmount = mkOr.MinOutputAmount
	return orders
}

type MsgDeleteOrder struct {
	MarketSymbol    string         `json:"market_symbol"`
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
	if len(strings.TrimSpace(m.MarketSymbol)) == 0 {
		return ErrInvalidMarket(m.MarketSymbol)
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
			IsBuy:           m.IsBuy,
		},
		OrderID: m.OrderID,
		PrevKey: m.PrevKey,
	}
}

type MsgAddLiquidity struct {
	Owner           sdk.AccAddress `json:"owner"`
	Stock           string         `json:"stock"`
	Money           string         `json:"money"`
	StockIn         sdk.Int        `json:"stock_in"`
	MoneyIn         sdk.Int        `json:"money_in"`
	To              sdk.AccAddress `json:"to"`
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
		return ErrInvalidToken("token is empty")
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
	Sender          sdk.AccAddress `json:"sender"`
	Stock           string         `json:"stock"`
	Money           string         `json:"money"`
	Amount          sdk.Int        `json:"amount"`
	To              sdk.AccAddress `json:"to"`
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
