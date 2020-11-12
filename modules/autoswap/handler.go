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
		case types.MsgCreateTradingPair:
			return handleMsgCreateTradingPair(ctx, k, msg)
		case types.MsgAddLiquidity:
			return handleMsgAddLiquidity(ctx, k, msg)
		case types.MsgRemoveLiquidity:
			return handleMsgRemoveLiquidity(ctx, k, msg)
		case types.MsgCreateOrder:
			return handleMsgCreateOrder(ctx, k, msg)
		case types.MsgCancelOrder:
			return handleMsgCancelOrder(ctx, k, msg)
		default:
			return dex.ErrUnknownRequest(types.ModuleName, msg)
		}
	}
}

func handleMsgCreateTradingPair(ctx sdk.Context, k keepers.Keeper, msg types.MsgCreateTradingPair) sdk.Result {
	panic("TODO")
}

func handleMsgAddLiquidity(ctx sdk.Context, k keepers.Keeper, msg types.MsgAddLiquidity) sdk.Result {
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	//todo: get trading pair, if not exist, return;
	var liquidity sdk.Int
	to := msg.To
	if to.Empty() {
		to = msg.Sender
	}
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
		liquidity, err = k.IPairKeeper.Mint(ctx, marKey, stockR, moneyR, to)
		if err != nil {
			return err.Result()
		}
	}
	infoDisplay := keepers.NewAddLiquidityInfo(msg, liquidity)
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
	to := msg.To
	if to.Empty() {
		to = msg.Sender
	}
	err = k.SendCoinsFromPoolToUser(ctx, to, sdk.NewCoins(sdk.NewCoin(msg.Stock, stockOut), sdk.NewCoin(msg.Money, moneyOut)))
	if err != nil {
		return err.Result()
	}
	infoDisplay := keepers.NewRemoveLiquidityInfo(msg, stockOut, moneyOut)
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

func handleMsgCreateOrder(ctx sdk.Context, k keepers.Keeper, msg types.MsgCreateOrder) sdk.Result {
	if err := k.AddLimitOrder(ctx, msg.GetOrder()); err != nil {
		return err.Result()
	}
	return sdk.Result{}
}
func handleMsgCancelOrder(ctx sdk.Context, k keepers.Keeper, msg types.MsgCancelOrder) sdk.Result {
	if err := k.DeleteOrder(ctx, msg); err != nil {
		return err.Result()
	}
	return sdk.Result{}
}

func fillMsgQueue(ctx sdk.Context, keeper Keeper, key string, msg interface{}) {
	if keeper.IsSubscribed(types.Topic) {
		msgqueue.FillMsgs(ctx, key, msg)
	}
}
