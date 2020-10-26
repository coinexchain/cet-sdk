package types

import (
	"math"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderBasic struct {
	MarketSymbol string
	IsOpenSwap   bool
	Sender       sdk.AccAddress
	Amount       int64
	IsBuy        bool
	IsLimitOrder bool
}

type Order struct {
	OrderBasic
	Price          int64
	PricePrecision int64
	OrderID        int64
	NextOrderID    int64
	PrevKey        [3]int64 `json:"-"`

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
		return or.actualPrice.Mul(sdk.NewDec(or.Amount)).Ceil().RoundInt()
	}
	return sdk.NewInt(or.Amount)
}

func (or *Order) ActualPrice() sdk.Dec {
	if !or.actualPrice.IsZero() {
		return or.actualPrice
	}
	return sdk.NewDec(or.Price).Quo(sdk.NewDec(int64(math.Pow10(int(or.PricePrecision)))))
}
