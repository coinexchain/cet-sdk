package autoswap

import (
	"fmt"

	"github.com/coinexchain/cet-sdk/msgqueue"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"
)

func NewInternalHandler(k *keepers.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgAutoSwapCreateTradingPair:
			return handleMsgCreateTradingPair(ctx, k, msg)
		case types.MsgCancelTradingPair:
			return handleMsgCancelTradingPair(ctx, k, msg)
		case types.MsgAddLiquidity:
			return handleMsgAddLiquidity(ctx, k, msg)
		case types.MsgRemoveLiquidity:
			return handleMsgRemoveLiquidity(ctx, k, msg)
		case types.MsgAutoSwapCreateOrder:
			return handleMsgCreateOrder(ctx, k, msg)
		case types.MsgAutoSwapCancelOrder:
			return handleMsgCancelOrder(ctx, k, msg)
		default:
			return dex.ErrUnknownRequest(types.ModuleName, msg)
		}
	}
}

func handleMsgCreateTradingPair(ctx sdk.Context, k *keepers.Keeper, msg types.MsgAutoSwapCreateTradingPair) sdk.Result {
	if k.ExpectedAssetKeeper.GetToken(ctx, msg.Stock) == nil {
		return types.ErrInvalidToken(fmt.Sprintf("token: %s not exist", msg.Stock)).Result()
	}
	if k.ExpectedAssetKeeper.GetToken(ctx, msg.Money) == nil {
		return types.ErrInvalidToken(fmt.Sprintf("token: %s not exist", msg.Money)).Result()
	}
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	info := k.IPairKeeper.GetPoolInfo(ctx, marKey)
	if info != nil {
		return types.ErrPairAlreadyExist().Result()
	}
	k.CreatePair(ctx, msg.Creator, marKey, msg.PricePrecision)
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

func handleMsgCancelTradingPair(ctx sdk.Context, k *keepers.Keeper, msg types.MsgCancelTradingPair) sdk.Result {
	return sdk.Result{}
}

func handleMsgAddLiquidity(ctx sdk.Context, k *keepers.Keeper, msg types.MsgAddLiquidity) sdk.Result {
	marKey := dex.GetSymbol(msg.Stock, msg.Money)
	info := k.GetPoolInfo(ctx, marKey)
	if info == nil {
		return types.ErrPairIsNotExist().Result()
	}
	var liquidity sdk.Int
	to := msg.To
	if to.Empty() {
		to = msg.Sender
	}
	stockR, moneyR := info.GetLiquidityAmountIn(msg.StockIn, msg.MoneyIn)
	err := k.SendCoinsFromUserToPool(ctx, msg.Sender, sdk.NewCoins(sdk.NewCoin(msg.Stock, stockR), sdk.NewCoin(msg.Money, moneyR)))
	if err != nil {
		return err.Result()
	}
	liquidity, err = k.IPairKeeper.Mint(ctx, marKey, stockR, moneyR, to)
	if err != nil {
		return err.Result()
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

func handleMsgRemoveLiquidity(ctx sdk.Context, k *keepers.Keeper, msg types.MsgRemoveLiquidity) sdk.Result {
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

func handleMsgCreateOrder(ctx sdk.Context, k *keepers.Keeper, msg types.MsgAutoSwapCreateOrder) sdk.Result {
	if err := k.AddLimitOrder(ctx, msg.GetOrder()); err != nil {
		return err.Result()
	}
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
func handleMsgCancelOrder(ctx sdk.Context, k *keepers.Keeper, msg types.MsgAutoSwapCancelOrder) sdk.Result {
	if err := k.DeleteOrder(ctx, msg); err != nil {
		return err.Result()
	}
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

func fillMsgQueue(ctx sdk.Context, keeper *Keeper, key string, msg interface{}) {
	if keeper.IsSubscribed(types.Topic) {
		msgqueue.FillMsgs(ctx, key, msg)
	}
}
