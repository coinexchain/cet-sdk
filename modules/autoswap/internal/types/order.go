package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderBasic struct {
	MarketSymbol string
	IsOpenSwap   bool
	Sender       sdk.AccAddress
	Amount       uint64
	IsBuy        bool
	IsLimitOrder bool
}

type Order struct {
	OrderBasic
	Price          uint64
	PricePrecision uint64
	OrderID        int64
	NextOrderID    int64
	PrevKey        [3]int64 `json:"-"`

	// cache
	stock string
	money string
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
