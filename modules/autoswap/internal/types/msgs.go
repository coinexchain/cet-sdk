package types

import (
	"fmt"
)

type MsgCreateLimitOrder struct {
	OrderBasic
	OrderID        uint64
	Price          uint64
	PricePrecision byte
	PrevKey        uint64
}

func (limit MsgCreateLimitOrder) String() string {
	content := fmt.Sprintf("Sender: %s, Price: %d, PricePrecision: %d,Amount: "+
		"%d, OrderID: %d\n", limit.Sender.String(), limit.Price, limit.PricePrecision, limit.Amount, limit.OrderID)
	return content
}

func (limit MsgCreateLimitOrder) ValidateBasic() bool {
	if limit.Sender.Empty() || limit.Price == 0 || limit.Amount == 0 {
		return false
	}
	return true
}

type MsgCreateMarketOrder struct {
	OrderBasic
}

func (mkOr MsgCreateMarketOrder) String() string {
	return fmt.Sprintf("Sender: %s, MarketSymbol: %s, Amount: %d, IsBuy: %v\n",
		mkOr.Sender.String(), mkOr.MarketSymbol, mkOr.Amount, mkOr.IsBuy)
}

func (mkOr MsgCreateMarketOrder) ValidateBasic() bool {
	if mkOr.Sender.Empty() || mkOr.Amount == 0 {
		return false
	}
	return true
}
