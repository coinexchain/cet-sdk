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
		default:
			return dex.ErrUnknownRequest(types.ModuleName, msg)
		}
	}
}

func handleMsgAddLiquidity(ctx sdk.Context, k keepers.Keeper, msg types.MsgAddLiquidity) sdk.Result {
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	info := k.GetPoolInfo(ctx, marKey, msg.IsOpenSwap)
	if info == nil {
		if !k.CreatePair(ctx, msg) {
			//todo:
			return sdk.Result{}
		}
		return sdk.Result{}
	}
	stockR, moneyR := info.GetLiquidityAmountIn(msg.StockIn, msg.MoneyIn)
	//transfer token
	err := k.Mint(ctx, marKey, msg.IsOpenSwap, stockR, moneyR, msg.To)
	if err != nil {
		return sdk.Result{}
	}
	return sdk.Result{}
}
