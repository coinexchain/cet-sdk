package keepers

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//var

type FactoryInterface interface {
	CreatePair(ctx sdk.Context, msg types.MsgCreatePair) bool
	QueryPair(ctx sdk.Context, marketSymbol string, isOpenSwap bool) *PoolInfo
}

type FactoryKeeper struct {
	storeKey   sdk.StoreKey
	poolKeeper PoolKeeper
}

func (f FactoryKeeper) CreatePair(ctx sdk.Context, msg types.MsgCreatePair) bool {
	symbol := dex.GetSymbol(msg.Stock, msg.Money)
	info := f.poolKeeper.GetPoolInfo(ctx, symbol, msg.IsOpenSwap)
	if info != nil {
		return false
	}
	p := &PoolInfo{
		symbol:                symbol,
		stockAmmReserve:       msg.StockIn,
		moneyAmmReserve:       msg.MoneyIn,
		stockOrderBookReserve: sdk.ZeroInt(),
		moneyOrderBookReserve: sdk.ZeroInt(),
		totalSupply:           sdk.ZeroInt(),
		kLast:                 sdk.ZeroInt(),
	}
	f.poolKeeper.SetPoolInfo(ctx, symbol, msg.IsOpenSwap, p)
	//vanity check in handler
	if msg.StockIn.IsPositive() {
		err := f.poolKeeper.Mint(ctx, symbol, msg.IsOpenSwap, msg.StockIn, msg.MoneyIn, msg.To)
		if err != nil {
			return false
		}
	}
	return true
}

func (f FactoryKeeper) QueryPair(ctx sdk.Context, marketSymbol string, isOpenSwap bool) *PoolInfo {
	return f.poolKeeper.GetPoolInfo(ctx, marketSymbol, isOpenSwap)
}
