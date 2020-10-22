package keepers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IPoolKeeper interface {
	SetPoolReserves(ctx sdk.Context, marketSymbol string, ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount int64)
	GetPoolReserves(ctx sdk.Context, marketSymbol string) (ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount int64)
	Mint(ctx sdk.Context, marketSymbol string, stockAmountIn, moneyAmountIn int64, to sdk.AccAddress) sdk.Error
	Burn(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error
	Skim(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error
	Sync(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error
}

type PoolKeeper struct {
	key sdk.StoreKey
	codec         *codec.Codec
}

func (p PoolKeeper) SetPoolReserves(ctx sdk.Context, marketSymbol string, ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount int64) {
	store := ctx.KVStore(p.key)
	info := PoolInfo{
		stockAmmReserve:       ammStockAmount,
		moneyAmmReserve:       ammMoneyAmount,
		stockOrderBookReserve: orderBookStockAmount,
		moneyOrderBookReserve: orderBookMoneyAmount,
	}
	bytes := p.codec.MustMarshalBinaryBare(info)
	store.Set(getPairKey(marketSymbol), bytes)
}

func (p PoolKeeper) GetPoolReserves(ctx sdk.Context, marketSymbol string) (ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount int64) {
	store := ctx.KVStore(p.key)
	info := PoolInfo{}
	bytes := store.Get(getPairKey(marketSymbol))
	if bytes == nil {
		return 0,0,0,0
	}
	p.codec.MustUnmarshalBinaryBare(bytes, &info)
	return info.stockAmmReserve, info.moneyAmmReserve, info.stockOrderBookReserve, info.moneyOrderBookReserve
}

func (p PoolKeeper) Mint(ctx sdk.Context, marketSymbol string, stockAmountIn, moneyAmountIn int64, to sdk.AccAddress) sdk.Error {
	return nil
}

func (p PoolKeeper) Burn(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error {
	return nil
}

func (p PoolKeeper) Skim(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error {
	return nil
}

func (p PoolKeeper) Sync(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error {
	return nil
}

func NewPoolKeeper(key sdk.StoreKey) *PoolKeeper {
	return &PoolKeeper{
		key: key,
	}
}

var _ IPoolKeeper = PoolKeeper{}

type PoolInfo struct {
	stockAmmReserve int64
	moneyAmmReserve int64
	stockOrderBookReserve int64
	moneyOrderBookReserve int64
}

