package types

import (
	"strings"

	"github.com/coinexchain/cet-sdk/modules/market"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Order struct {
	TradingPair          string         `json:"trading_pair"`
	Sequence             int64          `json:"sequence"`
	Identify             byte           `json:"identify"`
	Sender               sdk.AccAddress `json:"sender"`
	Price                sdk.Dec        `json:"price"`
	Quantity             int64          `json:"quantity"`
	Height               int64          `json:"height"`
	IsBuy                bool           `json:"side"`
	OrderIndexInOneBlock int32          `json:"order_index_in_one_block"`

	// These fields will change when order was filled/canceled.
	LeftStock int64 `json:"left_stock"`
	Freeze    int64 `json:"freeze"`
	DealStock int64 `json:"deal_stock"`
	DealMoney int64 `json:"deal_money"`

	// cache
	stock   string `json:"-"`
	money   string `json:"-"`
	orderID string `json:"-"`
}

func (or *Order) GetOrderID() string {
	if len(or.orderID) != 0 {
		return or.orderID
	}
	or.orderID = market.AssemblyOrderID(or.Sender.String(), uint64(or.Sequence), or.Identify)
	return or.orderID
}
func (or *Order) GetSide() byte {
	if or.IsBuy {
		return market.BUY
	} else {
		return market.SELL
	}
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
