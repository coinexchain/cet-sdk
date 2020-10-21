package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	BUY  = 0x01
	SELL = 0x02
)

type Order struct {
	MarketSymbol   string
	Sender         sdk.AccAddress
	Price          uint64
	PricePrecision byte
	Amount         uint64
	IsBuy          bool
	NextID         uint64
}

func (or Order) String() string {
	content := fmt.Sprintf("sender: %s, Price: %d, PricePrecision: %d,Amount: "+
		"%d, NextID: %d\n", or.Sender.String(), or.Price, or.PricePrecision, or.Amount, or.NextID)
	return content
}

func (or Order) ValidateBasic() bool {
	if or.Sender.Empty() || or.Price == 0 || or.Amount == 0 {
		return false
	}
	return true
}

func (or Order) StoreKey(orderID uint64) []byte {
	side := BUY
	if !or.IsBuy {
		side = SELL
	}
	return []byte(fmt.Sprintf("%d%s%d", side, or.MarketSymbol, orderID))
}
