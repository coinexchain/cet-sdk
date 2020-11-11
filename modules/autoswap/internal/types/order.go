package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Order struct {
	TradingPair string         `json:"trading_pair"`
	Sender      sdk.AccAddress `json:"sender"`
	OrderID     string         `json:"order_id"`
	Price       sdk.Dec        `json:"price"`
	Quantity    int64          `json:"quantity"`
	Height      int64          `json:"height"`
	IsBuy       bool           `json:"side"`

	// These fields will change when order was filled/canceled.
	LeftStock int64 `json:"left_stock"`
	Freeze    int64 `json:"freeze"`
	DealStock int64 `json:"deal_stock"`
	DealMoney int64 `json:"deal_money"`

	// cache
	stock string
	money string
}

func (or *Order) Stock() string {
	if or.stock != "" {
		return or.stock
	}
	or.parseStockAndMoney()
	return or.stock
}

func (or *Order) Money() string {
	if or.money != "" {
		return or.money
	}
	or.parseStockAndMoney()
	return or.money
}

func (or *Order) parseStockAndMoney() {
	symbols := strings.Split(or.TradingPair, "/")
	or.stock = symbols[0]
	or.money = symbols[1]
}

func (or *Order) ActualAmount() sdk.Int {
	if or.IsBuy {
		return or.Price.Mul(sdk.NewDec(or.LeftStock)).Ceil().RoundInt()
	}
	return sdk.NewInt(or.LeftStock)
}

func (or *Order) SetOrderID(orderID string) {
	or.OrderID = orderID
}
