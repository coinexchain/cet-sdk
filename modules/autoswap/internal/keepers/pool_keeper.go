package keepers

import (
	"fmt"
	"math/big"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IPoolKeeper interface {
	SetPoolInfo(ctx sdk.Context, marketSymbol string, info *PoolInfo)
	GetPoolInfo(ctx sdk.Context, marketSymbol string) *PoolInfo
	GetPoolInfos(ctx sdk.Context) (infos []PoolInfo)
	SetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress, liquidity sdk.Int)
	GetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress) sdk.Int
	ClearLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress)
	GetAllLiquidityInfos(ctx sdk.Context) (infos []LiquidityInfo)
	Mint(ctx sdk.Context, marketSymbol string, stockAmountIn, moneyAmountIn sdk.Int, to sdk.AccAddress) (sdk.Int, sdk.Error)
	Burn(ctx sdk.Context, marketSymbol string, from sdk.AccAddress, liquidity sdk.Int) (stockOut, moneyOut sdk.Int, err sdk.Error)
}

//todo: _mintFee is not support
type LiquidityInfo struct {
	Symbol    string         `json:"symbol"`
	Owner     sdk.AccAddress `json:"owner"`
	Liquidity sdk.Int        `json:"liquidity"`
}

type PoolKeeper struct {
	key   sdk.StoreKey
	codec *codec.Codec
	types.SupplyKeeper
}

func (p PoolKeeper) Mint(ctx sdk.Context, marketSymbol string, stockAmountIn, moneyAmountIn sdk.Int, to sdk.AccAddress) (sdk.Int, sdk.Error) {
	info := p.GetPoolInfo(ctx, marketSymbol)
	if info == nil {
		return sdk.ZeroInt(), types.ErrPairIsNotExist()
	}
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
		return sdk.ZeroInt(), types.ErrInvalidLiquidityAmount()
	}
	info.TotalSupply = info.TotalSupply.Add(liquidity)
	totalLiq := p.GetLiquidity(ctx, marketSymbol, to)
	totalLiq = totalLiq.Add(liquidity)
	p.SetLiquidity(ctx, marketSymbol, to, totalLiq)
	info.StockAmmReserve = info.StockAmmReserve.Add(stockAmountIn)
	info.MoneyAmmReserve = info.MoneyAmmReserve.Add(moneyAmountIn)
	p.SetPoolInfo(ctx, marketSymbol, info)
	return liquidity, nil
}

func (p PoolKeeper) Burn(ctx sdk.Context, marketSymbol string, from sdk.AccAddress, liquidity sdk.Int) (stockOut, moneyOut sdk.Int, err sdk.Error) {
	info := p.GetPoolInfo(ctx, marketSymbol)
	if info == nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrPairIsNotExist()
	}
	stockAmount := liquidity.Mul(info.StockAmmReserve).Quo(info.TotalSupply)
	moneyAmount := liquidity.Mul(info.MoneyAmmReserve).Quo(info.TotalSupply)
	info.StockAmmReserve = info.StockAmmReserve.Sub(stockAmount)
	info.MoneyAmmReserve = info.MoneyAmmReserve.Sub(moneyAmount)
	info.TotalSupply = info.TotalSupply.Sub(liquidity)
	l := p.GetLiquidity(ctx, marketSymbol, from)
	if l.LT(liquidity) {
		return sdk.ZeroInt(), sdk.ZeroInt(), types.ErrInvalidLiquidityAmount()
	}
	l = l.Sub(liquidity)
	if l.IsZero() {
		p.ClearLiquidity(ctx, marketSymbol, from)
	} else {
		p.SetLiquidity(ctx, marketSymbol, from, l)
	}
	p.SetPoolInfo(ctx, marketSymbol, info)
	return stockAmount, moneyAmount, nil
}

func (p PoolKeeper) ClearLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress) {
	store := ctx.KVStore(p.key)
	store.Delete(getLiquidityKey(marketSymbol, address))
}

func (p PoolKeeper) SetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress, liquidity sdk.Int) {
	store := ctx.KVStore(p.key)
	bytes := p.codec.MustMarshalBinaryBare(LiquidityInfo{
		Symbol:    marketSymbol,
		Owner:     address,
		Liquidity: liquidity,
	})
	store.Set(getLiquidityKey(marketSymbol, address), bytes)
}

func (p PoolKeeper) GetLiquidity(ctx sdk.Context, marketSymbol string, address sdk.AccAddress) sdk.Int {
	store := ctx.KVStore(p.key)
	info := LiquidityInfo{}
	bytes := store.Get(getLiquidityKey(marketSymbol, address))
	if bytes != nil {
		p.codec.MustUnmarshalBinaryBare(bytes, &info)
	}
	return info.Liquidity
}

func (p PoolKeeper) IterateAllLiquidityInfo(ctx sdk.Context, liquidityProc func(li LiquidityInfo)) {
	store := ctx.KVStore(p.key)
	iter := store.Iterator(PoolLiquidityKey, PoolLiquidityEndKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		li := &LiquidityInfo{}
		p.codec.MustUnmarshalBinaryBare(iter.Value(), li)
		liquidityProc(*li)
	}
}

func (p PoolKeeper) GetAllLiquidityInfos(ctx sdk.Context) (infos []LiquidityInfo) {
	proc := func(info LiquidityInfo) {
		infos = append(infos, info)
	}
	p.IterateAllLiquidityInfo(ctx, proc)
	return
}

func (p PoolKeeper) SetPoolInfo(ctx sdk.Context, marketSymbol string, info *PoolInfo) {
	store := ctx.KVStore(p.key)
	bytes := p.codec.MustMarshalBinaryBare(info)
	store.Set(getPairKey(marketSymbol), bytes)
}

func (p PoolKeeper) ClearPoolInfo(ctx sdk.Context, marketSymbol string) {
	store := ctx.KVStore(p.key)
	store.Delete(getPairKey(marketSymbol))
}

func (p PoolKeeper) IteratePoolInfo(ctx sdk.Context, poolInfoProc func(info *PoolInfo)) {
	store := ctx.KVStore(p.key)
	iter := store.Iterator(MarketKey, MarketEndKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		bi := &PoolInfo{}
		p.codec.MustUnmarshalBinaryBare(iter.Value(), bi)
		poolInfoProc(bi)
	}
}

func (p PoolKeeper) GetPoolInfos(ctx sdk.Context) (infos []PoolInfo) {
	proc := func(info *PoolInfo) {
		infos = append(infos, *info)
	}
	p.IteratePoolInfo(ctx, proc)
	return
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

var _ IPoolKeeper = PoolKeeper{}

type PoolInfo struct {
	Symbol                string  `json:"symbol"`
	StockAmmReserve       sdk.Int `json:"stock_amm_reserve"`
	MoneyAmmReserve       sdk.Int `json:"money_amm_reserve"`
	StockOrderBookReserve sdk.Int `json:"stock_order_book_reserve"`
	MoneyOrderBookReserve sdk.Int `json:"money_order_book_reserve"`
	TotalSupply           sdk.Int `json:"total_supply"`
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
	return fmt.Sprintf("Symbol:%v, StockAmmReserve: %s, "+
		"MoneyAmmReserve: %s, StockOrderBookReserve: %s, MoneyOrderBookReserve: %s, TotalSupply: %s\n",
		p.Symbol, p.StockAmmReserve, p.MoneyAmmReserve, p.StockOrderBookReserve,
		p.MoneyOrderBookReserve, p.TotalSupply)
}

type AddLiquidityInfo struct {
	Sender    sdk.AccAddress `json:"sender"`
	Stock     string         `json:"stock"`
	Money     string         `json:"money"`
	StockIn   sdk.Int        `json:"stock_in"`
	MoneyIn   sdk.Int        `json:"money_in"`
	To        sdk.AccAddress `json:"to"`
	Liquidity sdk.Int        `json:"liquidity"`
}

func NewAddLiquidityInfo(msg types.MsgAddLiquidity, liquidity sdk.Int) AddLiquidityInfo {
	to := msg.To
	if to.Empty() {
		to = msg.Sender
	}
	return AddLiquidityInfo{
		Sender:    msg.Sender,
		Stock:     msg.Stock,
		Money:     msg.Money,
		StockIn:   msg.StockIn,
		MoneyIn:   msg.MoneyIn,
		To:        to,
		Liquidity: liquidity,
	}
}

type RemoveLiquidityInfo struct {
	Sender   sdk.AccAddress `json:"sender"`
	Stock    string         `json:"stock"`
	Money    string         `json:"money"`
	Amount   sdk.Int        `json:"amount"`
	To       sdk.AccAddress `json:"to"`
	StockOut sdk.Int        `json:"stock_out"`
	MoneyOut sdk.Int        `json:"money_out"`
}

func NewRemoveLiquidityInfo(msg types.MsgRemoveLiquidity, stockOut, moneyOut sdk.Int) RemoveLiquidityInfo {
	to := msg.To
	if to.Empty() {
		to = msg.Sender
	}
	return RemoveLiquidityInfo{
		Sender:   msg.Sender,
		Stock:    msg.Stock,
		Money:    msg.Money,
		Amount:   msg.Amount,
		To:       to,
		StockOut: stockOut,
		MoneyOut: moneyOut,
	}
}
