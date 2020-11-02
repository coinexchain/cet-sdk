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
		default:
			return dex.ErrUnknownRequest(types.ModuleName, msg)
		}
	}
}

func handleMsgAddLiquidity(ctx sdk.Context, k keepers.Keeper, msg types.MsgAddLiquidity) sdk.Result {
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	info := k.IPairKeeper.GetPoolInfo(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen)
	if info == nil {
		if err := k.CreatePair(ctx, msg); err != nil {
			return err.Result()
		}
	} else {
		stockR, moneyR := info.GetLiquidityAmountIn(msg.StockIn, msg.MoneyIn)
		//transfer token
		err := k.IPairKeeper.Mint(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen, stockR, moneyR, msg.To)
		if err != nil {
			return err.Result()
		}
	}
	return sdk.Result{}
}

func handleMsgRemoveLiquidity(ctx sdk.Context, k keepers.Keeper, msg types.MsgRemoveLiquidity) sdk.Result {
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	stockOut, moneyOut, err := k.Burn(ctx, marKey, msg.IsSwapOpen, msg.IsOrderBookOpen, msg.Sender, msg.Amount)
	if err != nil {
		return err.Result()
	}
	if stockOut.LT(msg.AmountStockMin) || moneyOut.LT(msg.AmountMoneyMin) {
		return types.ErrAmountOutIsSmallerThanExpected().Result()
	}
	//transfer token
	//todo: move clear liquidity in burn
	return sdk.Result{}
}
