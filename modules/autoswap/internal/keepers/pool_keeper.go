package keepers

import (
	"math/big"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AMMPool = "auto_swap_amm_pool"
)

type IPoolKeeper interface {
	SetPoolInfo(ctx sdk.Context, marketSymbol string, isOpenSwap bool, info *PoolInfo)
	GetPoolInfo(ctx sdk.Context, marketSymbol string, isOpenSwap bool) *PoolInfo
	SetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress, liquidity sdk.Int)
	GetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress) sdk.Int
	ClearLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress)
	Mint(ctx sdk.Context, marketSymbol string, isOpenSwap bool, stockAmountIn, moneyAmountIn sdk.Int, to sdk.AccAddress) sdk.Error
	Burn(ctx sdk.Context, marketSymbol string, isOpenSwap bool, from sdk.AccAddress, liquidity sdk.Int) (stockOut, moneyOut sdk.Int, err sdk.Error)
}

//todo: _mintFee is not support
//todo: skim and sync no need
var FeeOn bool //todo: parameter it

type PoolKeeper struct {
	key   sdk.StoreKey
	codec *codec.Codec
	types.SupplyKeeper
}

func (p PoolKeeper) SetPairInfos(ctx sdk.Context, marketSymbol string, isOpenSwap bool, ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount sdk.Int) {
	store := ctx.KVStore(p.key)
	info := PoolInfo{
		stockAmmReserve:       ammStockAmount,
		moneyAmmReserve:       ammMoneyAmount,
		stockOrderBookReserve: orderBookStockAmount,
		moneyOrderBookReserve: orderBookMoneyAmount,
	}
	bytes := p.codec.MustMarshalBinaryBare(info)
	store.Set(getPairKey(marketSymbol, isOpenSwap), bytes)
}

func (p PoolKeeper) GetPairInfos(ctx sdk.Context, marketSymbol string, isOpenSwap bool) (ammStockAmount, ammMoneyAmount, orderBookStockAmount, orderBookMoneyAmount sdk.Int) {
	info := p.GetPoolInfo(ctx, marketSymbol, isOpenSwap)
	return info.stockAmmReserve, info.moneyAmmReserve, info.stockOrderBookReserve, info.moneyOrderBookReserve
}

func (p PoolKeeper) Mint(ctx sdk.Context, marketSymbol string, isOpenSwap bool, stockAmountIn, moneyAmountIn sdk.Int, to sdk.AccAddress) sdk.Error {
	info := p.GetPoolInfo(ctx, marketSymbol, isOpenSwap)
	liquidity := sdk.ZeroInt()
	if info.totalSupply.IsZero() {
		value, _ := (&big.Int{}).SetString(stockAmountIn.Mul(moneyAmountIn).String(), 10)
		liquidity = sdk.NewIntFromBigInt(value.Sqrt(value))
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
	p.SetPoolInfo(ctx, marketSymbol, isOpenSwap, info)
	return nil
}

//todo: param check
func (p PoolKeeper) Burn(ctx sdk.Context, marketSymbol string, isOpenSwap bool, from sdk.AccAddress, liquidity sdk.Int) (stockOut, moneyOut sdk.Int, err sdk.Error) {
	info := p.GetPoolInfo(ctx, marketSymbol, isOpenSwap)
	stockAmount := liquidity.Mul(info.stockAmmReserve).Quo(info.totalSupply)
	moneyAmount := liquidity.Mul(info.moneyAmmReserve).Quo(info.totalSupply)
	info.stockAmmReserve = info.stockAmmReserve.Sub(stockAmount)
	info.moneyAmmReserve = info.moneyAmmReserve.Sub(moneyAmount)
	info.totalSupply = info.totalSupply.Sub(liquidity)
	l := p.GetLiquidity(ctx, marketSymbol, from)
	if l.LT(liquidity) {
		return sdk.ZeroInt(), sdk.ZeroInt(), nil
	}
	p.SetLiquidity(ctx, marketSymbol, from, l.Sub(liquidity))
	if FeeOn {
		info.kLast = info.stockAmmReserve.Mul(info.moneyAmmReserve)
	}
	return stockAmount, moneyAmount, nil
}

func (p PoolKeeper) ClearLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress) {
	store := ctx.KVStore(p.key)
	store.Delete(getLiquidityKey(marketSymbol, address))
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

func (p PoolKeeper) SetPoolInfo(ctx sdk.Context, marketSymbol string, isOpenSwap bool, info *PoolInfo) {
	store := ctx.KVStore(p.key)
	bytes := p.codec.MustMarshalBinaryBare(info)
	store.Set(getPairKey(marketSymbol, isOpenSwap), bytes)
}

func (p PoolKeeper) GetPoolInfo(ctx sdk.Context, marketSymbol string, isOpenSwap bool) *PoolInfo {
	store := ctx.KVStore(p.key)
	info := PoolInfo{}
	bytes := store.Get(getPairKey(marketSymbol, isOpenSwap))
	if bytes == nil {
		return nil
	}
	p.codec.MustUnmarshalBinaryBare(bytes, &info)
	return &info
}

func NewPoolKeeper(key sdk.StoreKey, keeper types.SupplyKeeper) *PoolKeeper {
	return &PoolKeeper{
		key:          key,
		SupplyKeeper: keeper,
	}
}

var _ IPoolKeeper = PoolKeeper{}

type PoolInfo struct {
	symbol                string
	stockAmmReserve       sdk.Int
	moneyAmmReserve       sdk.Int
	stockOrderBookReserve sdk.Int
	moneyOrderBookReserve sdk.Int
	totalSupply           sdk.Int
	kLast                 sdk.Int
}

func (p PoolInfo) GetSymbol() string {
	return p.symbol
}

func (p PoolInfo) GetLiquidityAmountIn(amountStockIn, amountMoneyIn sdk.Int) (amountStockOut, amountMoneyOut sdk.Int) {
	if !p.moneyAmmReserve.IsZero() && !p.stockAmmReserve.IsZero() {
		stockRequired := amountMoneyIn.Mul(p.stockAmmReserve).Quo(p.moneyAmmReserve)
		if stockRequired.LT(amountStockIn) {
			return stockRequired, amountMoneyIn
		}
		return amountStockIn, amountStockIn.Mul(p.moneyAmmReserve).Quo(p.stockAmmReserve)
	}
	return sdk.ZeroInt(), sdk.ZeroInt()
}

func (p PoolInfo) GetTokensAmountOut(liquidity sdk.Int) (stockOut, moneyOut sdk.Int) {
	stockOut = liquidity.Mul(p.stockAmmReserve).Quo(p.totalSupply)
	moneyOut = liquidity.Mul(p.moneyAmmReserve).Quo(p.totalSupply)
	return
}

type PoolInfoDisplay struct {
	Symbol                  string  `json:"symbol"`
	StockReserveInAmm       sdk.Int `json:"stock_reserve_in_amm"`
	MoneyReserveInAmm       sdk.Int `json:"money_reserve_in_amm"`
	StockReserveInOrderBook sdk.Int `json:"stock_reserve_in_order_book"`
	MoneyReserveInOrderBook sdk.Int `json:"money_reserve_in_order_book"`
	TotalLiquidity          sdk.Int `json:"total_liquidity"`
}

func NewPoolInfoDisplay(info *PoolInfo) PoolInfoDisplay {
	return PoolInfoDisplay{
		Symbol:                  info.symbol,
		StockReserveInAmm:       info.stockAmmReserve,
		MoneyReserveInAmm:       info.moneyAmmReserve,
		StockReserveInOrderBook: info.stockOrderBookReserve,
		MoneyReserveInOrderBook: info.moneyOrderBookReserve,
		TotalLiquidity:          info.totalSupply,
	}
}
