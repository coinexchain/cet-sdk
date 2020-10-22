package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type OrderBasic struct {
	MarketSymbol string
	IsOpenSwap   bool
	Sender       sdk.AccAddress
	Amount       uint64
	IsBuy        bool
}

type Order struct {
	OrderBasic
	Price          uint64
	PricePrecision uint64
	OrderID        int64
	NextOrderID    int64
	PrevKey        [3]int64 `json:"-"`
}

func (or Order) HasPrevKey() bool {
	for _, v := range or.PrevKey {
		if v >= 0 {
			return true
		}
	}
	return false
}
