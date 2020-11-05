package keepers

import (
	"fmt"
	"math/big"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IPoolKeeper interface {
	SetPoolInfo(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, info *PoolInfo)
	GetPoolInfo(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool) *PoolInfo
	SetLiquidity(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, address sdk.AccAddress, liquidity sdk.Int)
	GetLiquidity(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, address sdk.AccAddress) sdk.Int
	ClearLiquidity(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, address sdk.AccAddress)
	Mint(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, stockAmountIn, moneyAmountIn sdk.Int, to sdk.AccAddress) sdk.Error
	Burn(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, from sdk.AccAddress, liquidity sdk.Int) (stockOut, moneyOut sdk.Int, err sdk.Error)
}

//todo: _mintFee is not support
var FeeOn bool //todo: parameter it

type PoolKeeper struct {
	key   sdk.StoreKey
	codec *codec.Codec
	types.SupplyKeeper
}

func (p PoolKeeper) Mint(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, stockAmountIn, moneyAmountIn sdk.Int, to sdk.AccAddress) sdk.Error {
	info := p.GetPoolInfo(ctx, marketSymbol, isOpenSwap, isOpenOrderBook)
	liquidity := sdk.ZeroInt()
	if info.TotalSupply.IsZero() {
		value, _ := (&big.Int{}).SetString(stockAmountIn.Mul(moneyAmountIn).String(), 10)
		liquidity = sdk.NewIntFromBigInt(value.Sqrt(value))
	} else {
		liquidity = stockAmountIn.Mul(info.TotalSupply).Quo(info.StockAmmReserve)
		another := moneyAmountIn.Mul(info.TotalSupply).Quo(info.MoneyAmmReserve)
		if liquidity.GT(another) {
			liquidity = another
		}
	}
	if !liquidity.IsPositive() {
		return types.ErrInvalidLiquidityAmount()
	}
	info.TotalSupply = info.TotalSupply.Add(liquidity)
	totalLiq := p.GetLiquidity(ctx, marketSymbol, isOpenSwap, isOpenOrderBook, to)
	totalLiq = totalLiq.Add(liquidity)
	p.SetLiquidity(ctx, marketSymbol, isOpenSwap, isOpenOrderBook, to, totalLiq)
	info.StockAmmReserve = info.StockAmmReserve.Add(stockAmountIn)
	info.MoneyAmmReserve = info.MoneyAmmReserve.Add(moneyAmountIn)
	if FeeOn {
		info.KLast = info.StockAmmReserve.Mul(info.MoneyAmmReserve)
	}
	p.SetPoolInfo(ctx, marketSymbol, isOpenSwap, isOpenOrderBook, info)
	return nil
}

func (p PoolKeeper) Burn(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, from sdk.AccAddress, liquidity sdk.Int) (stockOut, moneyOut sdk.Int, err sdk.Error) {
	info := p.GetPoolInfo(ctx, marketSymbol, isOpenSwap, isOpenOrderBook)
	if info == nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrPairIsNotExist()
	}
	stockAmount := liquidity.Mul(info.StockAmmReserve).Quo(info.TotalSupply)
	moneyAmount := liquidity.Mul(info.MoneyAmmReserve).Quo(info.TotalSupply)
	info.StockAmmReserve = info.StockAmmReserve.Sub(stockAmount)
	info.MoneyAmmReserve = info.MoneyAmmReserve.Sub(moneyAmount)
	info.TotalSupply = info.TotalSupply.Sub(liquidity)
	l := p.GetLiquidity(ctx, marketSymbol, isOpenSwap, isOpenOrderBook, from)
	if l.LT(liquidity) {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrInvalidLiquidityAmount()
	}
	p.SetLiquidity(ctx, marketSymbol, isOpenSwap, isOpenOrderBook, from, l.Sub(liquidity))
	if FeeOn {
		info.KLast = info.StockAmmReserve.Mul(info.MoneyAmmReserve)
	}
	p.SetPoolInfo(ctx, marketSymbol, isOpenSwap, isOpenOrderBook, info)
	return stockAmount, moneyAmount, nil
}

func (p PoolKeeper) ClearLiquidity(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, address sdk.AccAddress) {
	store := ctx.KVStore(p.key)
	store.Delete(getLiquidityKey(marketSymbol, isOpenSwap, isOpenOrderBook, address))
}

func (p PoolKeeper) SetLiquidity(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, address sdk.AccAddress, liquidity sdk.Int) {
	store := ctx.KVStore(p.key)
	bytes := p.codec.MustMarshalBinaryBare(liquidity)
	store.Set(getLiquidityKey(marketSymbol, isOpenSwap, isOpenOrderBook, address), bytes)
}

func (p PoolKeeper) GetLiquidity(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, address sdk.AccAddress) sdk.Int {
	store := ctx.KVStore(p.key)
	liquidity := sdk.ZeroInt()
	bytes := store.Get(getLiquidityKey(marketSymbol, isOpenSwap, isOpenOrderBook, address))
	if bytes != nil {
		p.codec.MustUnmarshalBinaryBare(bytes, &liquidity)
	}
	return liquidity
}

func (p PoolKeeper) SetPoolInfo(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool, info *PoolInfo) {
	store := ctx.KVStore(p.key)
	bytes := p.codec.MustMarshalBinaryBare(info)
	store.Set(getPairKey(marketSymbol, isOpenSwap, isOpenOrderBook), bytes)
}

func (p PoolKeeper) GetPoolInfo(ctx sdk.Context, marketSymbol string, isOpenSwap, isOpenOrderBook bool) *PoolInfo {
	store := ctx.KVStore(p.key)
	info := PoolInfo{}
	bytes := store.Get(getPairKey(marketSymbol, isOpenSwap, isOpenOrderBook))
	if bytes == nil {
		return nil
	}
	p.codec.MustUnmarshalBinaryBare(bytes, &info)
	return &info
}

var _ IPoolKeeper = PoolKeeper{}

type PoolInfo struct {
	Symbol                string
	IsSwapOpen            bool
	IsOrderBookOpen       bool
	StockAmmReserve       sdk.Int
	MoneyAmmReserve       sdk.Int
	StockOrderBookReserve sdk.Int
	MoneyOrderBookReserve sdk.Int
	TotalSupply           sdk.Int
	KLast                 sdk.Int
}

func NewPoolInfo(symbol string, stockAmmReserve sdk.Int, moneyAmmReserve sdk.Int, totalSupply sdk.Int) PoolInfo {
	poolInfo := PoolInfo{
		Symbol:          symbol,
		StockAmmReserve: stockAmmReserve,
		MoneyAmmReserve: moneyAmmReserve,
		TotalSupply:     totalSupply,
		KLast:           stockAmmReserve.Mul(moneyAmmReserve),
	}
	return poolInfo
}

func (p PoolInfo) GetSymbol() string {
	return p.Symbol
}

func (p PoolInfo) GetLiquidityAmountIn(amountStockIn, amountMoneyIn sdk.Int) (amountStockOut, amountMoneyOut sdk.Int) {
	if !p.MoneyAmmReserve.IsZero() && !p.StockAmmReserve.IsZero() {
		stockRequired := amountMoneyIn.Mul(p.StockAmmReserve).Quo(p.MoneyAmmReserve)
		if stockRequired.LT(amountStockIn) {
			return stockRequired, amountMoneyIn
		}
		return amountStockIn, amountStockIn.Mul(p.MoneyAmmReserve).Quo(p.StockAmmReserve)
	}
	return sdk.ZeroInt(), sdk.ZeroInt()
}

func (p PoolInfo) GetTokensAmountOut(liquidity sdk.Int) (stockOut, moneyOut sdk.Int) {
	stockOut = liquidity.Mul(p.StockAmmReserve).Quo(p.TotalSupply)
	moneyOut = liquidity.Mul(p.MoneyAmmReserve).Quo(p.TotalSupply)
	return
}

func (p PoolInfo) String() string {
	return fmt.Sprintf("Symbol:%v, IsSwapOpen: %v, IsOrderBookOpen: %v, StockAmmReserve: %s, "+
		"MoneyAmmReserve: %s, StockOrderBookReserve: %s, MoneyOrderBookReserve: %s, TotalSupply: %s, KLast: %s\n",
		p.Symbol, p.IsSwapOpen, p.IsOrderBookOpen, p.StockAmmReserve, p.MoneyAmmReserve, p.StockOrderBookReserve,
		p.MoneyOrderBookReserve, p.TotalSupply, p.KLast)
}

type PoolInfoDisplay struct {
	Symbol                  string  `json:"Symbol"`
	StockReserveInAmm       sdk.Int `json:"stock_reserve_in_amm"`
	MoneyReserveInAmm       sdk.Int `json:"money_reserve_in_amm"`
	StockReserveInOrderBook sdk.Int `json:"stock_reserve_in_order_book"`
	MoneyReserveInOrderBook sdk.Int `json:"money_reserve_in_order_book"`
	TotalLiquidity          sdk.Int `json:"total_liquidity"`
}

func NewPoolInfoDisplay(info *PoolInfo) PoolInfoDisplay {
	return PoolInfoDisplay{
		Symbol:                  info.Symbol,
		StockReserveInAmm:       info.StockAmmReserve,
		MoneyReserveInAmm:       info.MoneyAmmReserve,
		StockReserveInOrderBook: info.StockOrderBookReserve,
		MoneyReserveInOrderBook: info.MoneyOrderBookReserve,
		TotalLiquidity:          info.TotalSupply,
	}
}
