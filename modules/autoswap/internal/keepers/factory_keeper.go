package keepers

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//var

type FactoryInterface interface {
	CreatePair(ctx sdk.Context, msg types.MsgAddLiquidity) (sdk.Int, sdk.Error)
	QueryPair(ctx sdk.Context, marketSymbol string) *PoolInfo
}

type FactoryKeeper struct {
	storeKey   sdk.StoreKey
	poolKeeper PoolKeeper
}

func (f FactoryKeeper) CreatePair(ctx sdk.Context, msg types.MsgAddLiquidity) (sdk.Int, sdk.Error) {
	symbol := dex.GetSymbol(msg.Stock, msg.Money)
	info := f.poolKeeper.GetPoolInfo(ctx, symbol)
	if info != nil {
		return sdk.ZeroInt(), types.ErrPairAlreadyExist()
	}
	p := &PoolInfo{
		Symbol:                symbol,
		StockAmmReserve:       msg.StockIn,
		MoneyAmmReserve:       msg.MoneyIn,
		StockOrderBookReserve: sdk.ZeroInt(),
		MoneyOrderBookReserve: sdk.ZeroInt(),
		TotalSupply:           sdk.ZeroInt(),
	}
	f.poolKeeper.SetPoolInfo(ctx, symbol, p)
	to := msg.To
	if to.Empty() {
		to = msg.Sender
	}
	liquidity, err := f.poolKeeper.Mint(ctx, symbol, msg.StockIn, msg.MoneyIn, to)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	return liquidity, nil
}

func (f FactoryKeeper) QueryPair(ctx sdk.Context, marketSymbol string) *PoolInfo {
	return f.poolKeeper.GetPoolInfo(ctx, marketSymbol)
}
