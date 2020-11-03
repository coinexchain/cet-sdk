package keepers

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//var

type FactoryInterface interface {
	CreatePair(ctx sdk.Context, msg types.MsgAddLiquidity) sdk.Error
	QueryPair(ctx sdk.Context, marketSymbol string, isSwapOpen bool, isOrderBookOpen bool) *PoolInfo
}

type FactoryKeeper struct {
	storeKey   sdk.StoreKey
	poolKeeper PoolKeeper
}

func (f FactoryKeeper) CreatePair(ctx sdk.Context, msg types.MsgAddLiquidity) sdk.Error {
	symbol := dex.GetSymbol(msg.Stock, msg.Money)
	info := f.poolKeeper.GetPoolInfo(ctx, symbol, msg.IsSwapOpen, msg.IsOrderBookOpen)
	if info != nil {
		return types.ErrPairAlreadyExist()
	}
	p := &PoolInfo{
		Symbol:                symbol,
		StockAmmReserve:       msg.StockIn,
		MoneyAmmReserve:       msg.MoneyIn,
		StockOrderBookReserve: sdk.ZeroInt(),
		MoneyOrderBookReserve: sdk.ZeroInt(),
		TotalSupply:           sdk.ZeroInt(),
		KLast:                 sdk.ZeroInt(),
	}
	f.poolKeeper.SetPoolInfo(ctx, symbol, msg.IsSwapOpen, msg.IsOrderBookOpen, p)
	//vanity check in handler
	if msg.StockIn.IsPositive() {
		err := f.poolKeeper.Mint(ctx, symbol, msg.IsSwapOpen, msg.IsOrderBookOpen, msg.StockIn, msg.MoneyIn, msg.To)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f FactoryKeeper) QueryPair(ctx sdk.Context, marketSymbol string, isSwapOpen bool, isOrderBookOpen bool) *PoolInfo {
	return f.poolKeeper.GetPoolInfo(ctx, marketSymbol, isSwapOpen, isOrderBookOpen)
}
