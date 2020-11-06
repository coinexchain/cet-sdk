package autoswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"
)

func NewHandler(k keepers.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgAddLiquidity:
			return handleMsgAddLiquidity(ctx, k, msg)
		case types.MsgRemoveLiquidity:
			return handleMsgRemoveLiquidity(ctx, k, msg)
		case types.MsgCreateLimitOrder:
			return handleMsgCreateLimitOrder(ctx, k, msg)
		case types.MsgSwapTokens:
			return handlerMsgSwapTokens(ctx, k, msg)
		case types.MsgDeleteOrder:
			return handlerMsgDeleteOrder(ctx, k, msg)
		default:
			return dex.ErrUnknownRequest(types.ModuleName, msg)
		}
	}
}

func handleMsgAddLiquidity(ctx sdk.Context, k keepers.Keeper, msg types.MsgAddLiquidity) sdk.Result {
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	info := k.IPairKeeper.GetPoolInfo(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen)
	if info == nil {
		err := k.SendCoinsFromUserToPool(ctx, msg.Owner, sdk.NewCoins(sdk.NewCoin(msg.Stock, msg.StockIn), sdk.NewCoin(msg.Money, msg.MoneyIn)))
		if err != nil {
			return err.Result()
		}
		if err := k.CreatePair(ctx, msg); err != nil {
			return err.Result()
		}
	} else {
		stockR, moneyR := info.GetLiquidityAmountIn(msg.StockIn, msg.MoneyIn)
		err := k.SendCoinsFromUserToPool(ctx, msg.Owner, sdk.NewCoins(sdk.NewCoin(msg.Stock, stockR), sdk.NewCoin(msg.Money, moneyR)))
		if err != nil {
			return err.Result()
		}
		err = k.IPairKeeper.Mint(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen, stockR, moneyR, msg.To)
		if err != nil {
			return err.Result()
		}
		kLast := info.KLast
		info = k.IPairKeeper.GetPoolInfo(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen)
		err = k.AllocateFeeToValidator(ctx, &kLast, info)
		if err != nil {
			return err.Result()
		}
	}
	return sdk.Result{}
}

func handleMsgRemoveLiquidity(ctx sdk.Context, k keepers.Keeper, msg types.MsgRemoveLiquidity) sdk.Result {
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	info := k.IPairKeeper.GetPoolInfo(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen)
	stockOut, moneyOut, err := k.Burn(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen, msg.Sender, msg.Amount)
	if err != nil {
		return err.Result()
	}
	if stockOut.LT(msg.AmountStockMin) {
		return types.ErrAmountOutIsSmallerThanExpected(msg.AmountStockMin, stockOut).Result()
	}
	if moneyOut.LT(msg.AmountMoneyMin) {
		return types.ErrAmountOutIsSmallerThanExpected(msg.AmountMoneyMin, moneyOut).Result()
	}

	kLast := info.KLast
	info = k.IPairKeeper.GetPoolInfo(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen)
	err = k.AllocateFeeToValidator(ctx, &kLast, info)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{}
}

func handleMsgCreateLimitOrder(ctx sdk.Context, k keepers.Keeper, msg types.MsgCreateLimitOrder) sdk.Result {
	if err := k.AddLimitOrder(ctx, msg.OrderInfo()); err != nil {
		return err.Result()
	}
	return sdk.Result{}
}

func handlerMsgSwapTokens(ctx sdk.Context, k keepers.Keeper, msg types.MsgSwapTokens) sdk.Result {
	orders := msg.GetOrderInfos()
	for i := 0; i < len(orders); i++ {
		outAmount, err := k.AddMarketOrder(ctx, orders[i])
		if err != nil {
			return err.Result()
		}
		if i < len(orders)-1 {
			orders[i+1].Amount = outAmount
		}
	}
	return sdk.Result{}
}

func handlerMsgDeleteOrder(ctx sdk.Context, k keepers.Keeper, msg types.MsgDeleteOrder) sdk.Result {
	if err := k.DeleteOrder(ctx, &msg); err != nil {
		return err.Result()
	}
	return sdk.Result{}
}
