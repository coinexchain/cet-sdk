package keepers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//var

type FactoryInterface interface {
	CreatePair(ctx sdk.Context, owner sdk.AccAddress, symbol string, pricePrecision byte)
	QueryPair(ctx sdk.Context, marketSymbol string) *PoolInfo
}

type FactoryKeeper struct {
	storeKey   sdk.StoreKey
	poolKeeper PoolKeeper
}

func (f FactoryKeeper) CreatePair(ctx sdk.Context, owner sdk.AccAddress, symbol string, pricePrecision byte) {
	p := &PoolInfo{
		Owner:                 owner,
		Symbol:                symbol,
		StockAmmReserve:       sdk.ZeroInt(),
		MoneyAmmReserve:       sdk.ZeroInt(),
		StockOrderBookReserve: sdk.ZeroInt(),
		MoneyOrderBookReserve: sdk.ZeroInt(),
		TotalSupply:           sdk.ZeroInt(),
		PricePrecision:        pricePrecision,
	}
	f.poolKeeper.SetPoolInfo(ctx, symbol, p)
}

func (f FactoryKeeper) QueryPair(ctx sdk.Context, marketSymbol string) *PoolInfo {
	return f.poolKeeper.GetPoolInfo(ctx, marketSymbol)
}
