package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderBasic struct {
	MarketSymbol    string         `json:"market_symbol"`
	IsOpenSwap      bool           `json:"is_open_swap"`
	IsOpenOrderBook bool           `json:"is_open_order_book"`
	Sender          sdk.AccAddress `json:"sender"`
	IsBuy           bool           `json:"is_buy"`
	IsLimitOrder    bool           `json:"is_limit_order"`

	// if the order is market_order, the amount is the actual input amount with special token(
	// ie: sell order, amount = stockTokenAmount, buy order = moneyTokenAmount)
	// if the order is limit_order, the amount is the stock amount and orderActualAmount will be calculated
	// (ie: buyActualAmount = price * amount, sellActualAmount = amount)
	Amount sdk.Int `json:"amount"`
}

type Order struct {
	OrderBasic
	Price           sdk.Dec
	OrderID         int64
	NextOrderID     int64
	PrevKey         [3]int64 `json:"-"`
	MinOutputAmount sdk.Int  `json:"-"`

	// cache
	stock       string
	money       string
	actualPrice sdk.Dec
}

func (or Order) HasPrevKey() bool {
	for _, v := range or.PrevKey {
		if v >= 0 {
			return true
		}
	}
	return false
}

func (or *Order) Stock() string {
	if or.stock != "" {
		return or.stock
	}
	return or.stock
}

func (or *Order) parseStockAndMoney() {
	symbols := strings.Split(or.MarketSymbol, "/")
	or.stock = symbols[0]
	or.money = symbols[1]
}

func (or *Order) Money() string {
	if or.money != "" {
		return or.money
	}
	or.parseStockAndMoney()
	return or.money
}

func (or *Order) ActualAmount() sdk.Int {
	if or.IsBuy {
		return or.actualPrice.Mul(sdk.NewDecFromInt(or.Amount)).Ceil().RoundInt()
	}
	return or.Amount
}
