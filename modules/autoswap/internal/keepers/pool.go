package keepers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/big"
)

type IPoolKeeper interface {
	SetPoolReserves(ctx sdk.Context, marketSymbol string, ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount sdk.Int)
	GetPoolReserves(ctx sdk.Context, marketSymbol string) (ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount sdk.Int)
	GetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress) sdk.Int
	Mint(ctx sdk.Context, marketSymbol string, stockAmountIn, moneyAmountIn sdk.Int, to sdk.AccAddress) sdk.Error
	Burn(ctx sdk.Context, marketSymbol string, liquidity sdk.Int, to sdk.AccAddress) (stockOut, moneyOut sdk.Int, err sdk.Error)
	Skim(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error
	Sync(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error
}

//todo: _mintFee is not support
//todo: skim and sync no need
var FeeOn bool //todo: parameter it

type PoolKeeper struct {
	key   sdk.StoreKey
	codec *codec.Codec
}

func (p PoolKeeper) SetPoolReserves(ctx sdk.Context, marketSymbol string, ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount sdk.Int) {
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

func (p PoolKeeper) GetPoolReserves(ctx sdk.Context, marketSymbol string) (ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount sdk.Int) {
	info := p.GetPoolInfo(ctx, marketSymbol)
	return info.stockAmmReserve, info.moneyAmmReserve, info.stockOrderBookReserve, info.moneyOrderBookReserve
}

func (p PoolKeeper) Mint(ctx sdk.Context, marketSymbol string, stockAmountIn, moneyAmountIn sdk.Int, to sdk.AccAddress) sdk.Error {
	info := p.GetPoolInfo(ctx, marketSymbol)
	liquidity := sdk.ZeroInt()
	if info.totalSupply.IsZero() {
		value, _ := (&big.Int{}).SetString(stockAmountIn.Mul(moneyAmountIn).String(), 10)
		liquidity = value.Sqrt(value)
	} else {
		liquidity = stockAmountIn.Mul(info.totalSupply).Quo(info.stockAmmReserve)
		another := moneyAmountIn.Mul(info.totalSupply).Quo(info.moneyAmmReserve)
		if liquidity.GT(another) {
			liquidity = another
		}
	}
	if !liquidity.IsPositive() {
		//todo: error type
		return nil
	}
	info.totalSupply = info.totalSupply.Add(liquidity)
	totalLiq := p.GetLiquidity(ctx, marketSymbol, to)
	totalLiq = totalLiq.Add(liquidity)
	p.SetLiquidity(ctx, marketSymbol, to, totalLiq)
	info.stockAmmReserve = info.stockAmmReserve.Add(stockAmountIn)
	info.moneyAmmReserve = info.moneyAmmReserve.Add(moneyAmountIn)
	if FeeOn {
		info.kLast = info.stockAmmReserve.Mul(info.moneyAmmReserve)
	}
	p.SetPoolInfo(ctx, marketSymbol, *info)
	return nil
}

//todo: param check
func (p PoolKeeper) Burn(ctx sdk.Context, marketSymbol string, liquidity sdk.Int, to sdk.AccAddress) (stockOut, moneyOut sdk.Int, err sdk.Error) {
	info := p.GetPoolInfo(ctx, marketSymbol)
	stockAmount := liquidity.Mul(info.stockAmmReserve).Quo(info.totalSupply)
	moneyAmount := liquidity.Mul(info.moneyAmmReserve).Quo(info.totalSupply)
	info.stockAmmReserve = info.stockAmmReserve.Sub(stockAmount)
	info.moneyAmmReserve = info.moneyAmmReserve.Sub(moneyAmount)
	info.totalSupply = info.totalSupply.Sub(liquidity)
	if FeeOn {
		info.kLast = info.stockAmmReserve.Mul(info.moneyAmmReserve)
	}
	return stockAmount, moneyAmount, nil
}

func (p PoolKeeper) Skim(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error {
	return nil
}

func (p PoolKeeper) Sync(ctx sdk.Context, marketSymbol string, to sdk.AccAddress) sdk.Error {
	return nil
}

func (p PoolKeeper) SetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress, liquidity sdk.Int) {
	store := ctx.KVStore(p.key)
	bytes := p.codec.MustMarshalBinaryBare(liquidity)
	store.Set(getLiquidityKey(marketSymbol, address), bytes)
}

func (p PoolKeeper) GetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress) sdk.Int {
	store := ctx.KVStore(p.key)
	liquidity := sdk.ZeroInt()
	bytes := store.Get(getLiquidityKey(marketSymbol, address))
	if bytes != nil {
		p.codec.MustUnmarshalBinaryBare(bytes, &liquidity)
	}
	return liquidity
}

func (p PoolKeeper) SetPoolInfo(ctx sdk.Context, marketSymbol string, info PoolInfo) *PoolInfo {
	store := ctx.KVStore(p.key)
	bytes := p.codec.MustMarshalBinaryBare(info)
	store.Set(getPairKey(marketSymbol), bytes)
}

func (p PoolKeeper) GetPoolInfo(ctx sdk.Context, marketSymbol string) *PoolInfo {
	store := ctx.KVStore(p.key)
	info := PoolInfo{}
	bytes := store.Get(getPairKey(marketSymbol))
	if bytes == nil {
		return nil
	}
	p.codec.MustUnmarshalBinaryBare(bytes, &info)
	return &info
}

func NewPoolKeeper(key sdk.StoreKey) *PoolKeeper {
	return &PoolKeeper{
		key: key,
	}
}

var _ IPoolKeeper = PoolKeeper{}

type PoolInfo struct {
	stockAmmReserve       sdk.Int
	moneyAmmReserve       sdk.Int
	stockOrderBookReserve sdk.Int
	moneyOrderBookReserve sdk.Int
	totalSupply           sdk.Int
	kLast                 sdk.Int
}
