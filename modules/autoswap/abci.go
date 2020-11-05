package autoswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, keeper Keeper) {
	pairs := keeper.GetPairList()
	for pair := range pairs {
		info := keeper.GetPoolInfo(ctx, pair.Symbol, pair.IsSwapOpen, pair.IsOrderBookOpen)
		_ = keeper.AllocateFeeToValidator(ctx, &info.KLast, info)
		info.KLast = info.StockAmmReserve.Mul(info.MoneyAmmReserve)
		keeper.SetPoolInfo(ctx, pair.Symbol, pair.IsSwapOpen, pair.IsOrderBookOpen, info)
	}
	keeper.ClearPairList()
}
