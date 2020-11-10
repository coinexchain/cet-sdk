package autoswap

import (
	"github.com/coinexchain/cet-sdk/msgqueue"
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
	//todo: get trading pair, if not exist, return;
	var liquidity sdk.Int
	info := k.IPairKeeper.GetPoolInfo(ctx, marKey)
	if info == nil {
		err := k.SendCoinsFromUserToPool(ctx, msg.Sender, sdk.NewCoins(sdk.NewCoin(msg.Stock, msg.StockIn), sdk.NewCoin(msg.Money, msg.MoneyIn)))
		if err != nil {
			return err.Result()
		}
		if liquidity, err = k.CreatePair(ctx, msg); err != nil {
			return err.Result()
		}
	} else {
		stockR, moneyR := info.GetLiquidityAmountIn(msg.StockIn, msg.MoneyIn)
		err := k.SendCoinsFromUserToPool(ctx, msg.Sender, sdk.NewCoins(sdk.NewCoin(msg.Stock, stockR), sdk.NewCoin(msg.Money, moneyR)))
		if err != nil {
			return err.Result()
		}
		liquidity, err = k.IPairKeeper.Mint(ctx, marKey, stockR, moneyR, msg.To)
		if err != nil {
			return err.Result()
		}
	}
	infoDisplay := keepers.AddLiquidityInfo{
		Sender:    msg.Sender,
		Stock:     msg.Stock,
		Money:     msg.Money,
		StockIn:   msg.StockIn,
		MoneyIn:   msg.MoneyIn,
		To:        msg.To,
		Liquidity: liquidity,
	}
	fillMsgQueue(ctx, k, KafkaAddLiquidity, infoDisplay)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeKeyAddLiquidity,
			sdk.NewAttribute(AttributeSymbol, marKey)),
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

func handleMsgRemoveLiquidity(ctx sdk.Context, k keepers.Keeper, msg types.MsgRemoveLiquidity) sdk.Result {
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	//todo: get trading pair, if not exist, return;
	stockOut, moneyOut, err := k.Burn(ctx, marKey, msg.Sender, msg.Amount)
	if err != nil {
		return err.Result()
	}
	err = k.SendCoinsFromPoolToUser(ctx, msg.Sender, sdk.NewCoins(sdk.NewCoin(msg.Stock, stockOut), sdk.NewCoin(msg.Money, moneyOut)))
	if err != nil {
		return err.Result()
	}
	infoDisplay := keepers.RemoveLiquidityInfo{
		Sender:   msg.Sender,
		Stock:    msg.Stock,
		Money:    msg.Money,
		Amount:   msg.Amount,
		To:       msg.To,
		StockOut: stockOut,
		MoneyOut: moneyOut,
	}
	fillMsgQueue(ctx, k, KafkaRemoveLiquidity, infoDisplay)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeKeyRemoveLiquidity,
			sdk.NewAttribute(AttributeSymbol, marKey)),
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
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

func fillMsgQueue(ctx sdk.Context, keeper Keeper, key string, msg interface{}) {
	if keeper.IsSubscribed(types.Topic) {
		msgqueue.FillMsgs(ctx, key, msg)
	}
}
