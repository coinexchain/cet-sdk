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
	OrderID        uint64
}
