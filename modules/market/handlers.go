package market

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/market/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/market/internal/types"
	"github.com/coinexchain/cet-sdk/msgqueue"
	dex "github.com/coinexchain/cet-sdk/types"
)

func NewHandler(k keepers.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgCreateTradingPair:
			return handleMsgCreateTradingPair(ctx, msg, k)
		case types.MsgCreateOrder:
			return handleMsgCreateOrder(ctx, msg, k)
		case types.MsgCancelOrder:
			return handleMsgCancelOrder(ctx, msg, k)
		case types.MsgCancelTradingPair:
			return handleMsgCancelTradingPair(ctx, msg, k)
		case types.MsgModifyPricePrecision:
			return handleMsgModifyPricePrecision(ctx, msg, k)
		default:
			return dex.ErrUnknownRequest(ModuleName, msg)
		}
	}
}

func handleMsgCreateTradingPair(ctx sdk.Context, msg types.MsgCreateTradingPair, keeper keepers.Keeper) sdk.Result {
	if err := checkMsgCreateTradingPair(ctx, msg, keeper); err != nil {
		return err.Result()
	}

	var orderPrecision byte
	if msg.OrderPrecision <= types.MaxOrderPrecision {
		orderPrecision = msg.OrderPrecision
	}
	info := types.MarketInfo{
		Stock:             msg.Stock,
		Money:             msg.Money,
		PricePrecision:    msg.PricePrecision,
		LastExecutedPrice: sdk.ZeroDec(),
		OrderPrecision:    orderPrecision,
	}

	if err := keeper.SetMarket(ctx, info); err != nil {
		// only MarshalBinaryBare can cause error here, which is impossible in production
		return err.Result()
	}

	param := keeper.GetParams(ctx)
	if err := keeper.SubtractFeeAndCollectFee(ctx, msg.Creator, param.CreateMarketFee); err != nil {
		// CreateMarketFee has been checked with HasCoins in checkMsgCreateTradingPair
		// this clause will not execute in production
		return err.Result()
	}

	sendCreateMarketMsg(ctx, keeper, &msg)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeKeyCreateTradingPair,
			sdk.NewAttribute(AttributeKeyTradingPair, msg.GetSymbol()),
			sdk.NewAttribute(AttributeKeyStock, msg.Stock),
			sdk.NewAttribute(AttributeKeyMoney, msg.Money),
			sdk.NewAttribute(AttributeKeyPricePrecision, strconv.Itoa(int(info.PricePrecision))),
			sdk.NewAttribute(AttributeKeyLastExecutePrice, info.LastExecutedPrice.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Creator.String()),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

func checkMsgCreateTradingPair(ctx sdk.Context, msg types.MsgCreateTradingPair, keeper keepers.Keeper) sdk.Error {
	if _, err := keeper.GetMarketInfo(ctx, msg.GetSymbol()); err == nil {
		return types.ErrRepeatTradingPair()
	}

	if !keeper.IsTokenExists(ctx, msg.Money) || !keeper.IsTokenExists(ctx, msg.Stock) {
		return types.ErrTokenNoExist()
	}

	if !keeper.IsTokenIssuer(ctx, msg.Stock, msg.Creator) {
		return types.ErrInvalidTokenIssuer()
	}

	marketParams := keeper.GetParams(ctx)
	if !keeper.HasCoins(ctx, msg.Creator, dex.NewCetCoins(marketParams.CreateMarketFee)) {
		return types.ErrInsufficientCoins()
	}

	return nil
}

func sendCreateMarketMsg(ctx sdk.Context, keeper keepers.Keeper, market *types.MsgCreateTradingPair) {
	if keeper.IsSubScribed(types.Topic) {
		msgqueue.FillMsgs(ctx, types.CreateMarketInfoKey, market)
	}
}

type ParamOfCommissionMsg struct {
	amountOfMoney sdk.Dec
	amountOfStock sdk.Dec
	stock         string
	money         string
}

func CalCommission(ctx sdk.Context, keeper keepers.QueryMarketInfoAndParams, msg ParamOfCommissionMsg) (int64, sdk.Error) {
	marketParams := keeper.GetParams(ctx)
	volume := keeper.GetMarketVolume(ctx, msg.stock, msg.money, msg.amountOfStock, msg.amountOfMoney)
	rate := sdk.NewDec(marketParams.MarketFeeRate).QuoInt64(int64(math.Pow10(types.DefaultMarketFeeRatePrecision)))
	commission := volume.Mul(rate).Ceil().RoundInt64()
	if commission > types.MaxOrderAmount {
		return 0, types.ErrInvalidOrderAmount("The frozen fee is too large")
	}
	if commission < marketParams.MarketFeeMin {
		commission = marketParams.MarketFeeMin
	}
	return commission, nil
}

func calOrderCommission(ctx sdk.Context, keeper keepers.QueryMarketInfoAndParams, msg types.MsgCreateOrder) (int64, sdk.Error) {
	moneyAmount, err := calculateAmount(msg.Price, msg.Quantity, msg.PricePrecision)
	if err != nil {
		return 0, types.ErrInvalidOrderAmount(err.Error())
	}
	stock, money := SplitSymbol(msg.TradingPair)
	commissionMsg := ParamOfCommissionMsg{
		amountOfMoney: moneyAmount,
		amountOfStock: sdk.NewDec(msg.Quantity),
		stock:         stock,
		money:         money,
	}
	return CalCommission(ctx, keeper, commissionMsg)
}

func calFeatureFeeForExistBlocks(msg types.MsgCreateOrder, marketParam types.Params) int64 {
	if msg.TimeInForce == types.IOC {
		return 0
	}
	if msg.ExistBlocks < marketParam.GTEOrderLifetime {
		return 0
	}
	fee := sdk.NewInt(msg.ExistBlocks - marketParam.GTEOrderLifetime).
		MulRaw(marketParam.GTEOrderFeatureFeeByBlocks)
	if fee.GT(sdk.NewInt(types.MaxOrderAmount)) {
		return types.MaxOrderAmount
	}
	return fee.Int64()
}

func handleFeeForCreateOrder(ctx sdk.Context, keeper keepers.Keeper, amount int64, denom string,
	sender sdk.AccAddress, frozenFee, featureFee int64) sdk.Error {
	coin := sdk.NewCoin(denom, sdk.NewInt(amount))
	if err := keeper.FreezeCoins(ctx, sender, sdk.Coins{coin}); err != nil {
		return err
	}
	if frozenFee != 0 {
		if err := keeper.FreezeCoins(ctx, sender, dex.NewCetCoins(frozenFee)); err != nil {
			return err
		}
	}
	if featureFee != 0 {
		if err := keeper.FreezeCoins(ctx, sender, dex.NewCetCoins(featureFee)); err != nil {
			return err
		}
	}
	return nil
}

func sendCreateOrderMsg(ctx sdk.Context, keeper keepers.Keeper, order types.Order) {
	if keeper.IsSubScribed(types.Topic) {
		// send msg to kafka
		createOrderInfo := types.CreateOrderInfo{
			OrderID:          order.OrderID(),
			Sender:           order.Sender.String(),
			TradingPair:      order.TradingPair,
			OrderType:        order.OrderType,
			Price:            order.Price,
			Quantity:         order.Quantity,
			Side:             order.Side,
			TimeInForce:      order.TimeInForce,
			Height:           order.Height,
			FrozenCommission: order.FrozenCommission,
			FrozenFeatureFee: order.FrozenFeatureFee,
			Freeze:           order.Freeze,
		}
		msgqueue.FillMsgs(ctx, types.CreateOrderInfoKey, createOrderInfo)
	}
}

func getDenomAndOrderAmount(msg types.MsgCreateOrder) (string, int64, sdk.Error) {
	stock, money := SplitSymbol(msg.TradingPair)
	denom := stock
	amount := msg.Quantity
	if msg.Side == types.BUY {
		denom = money
		tmpAmount, err := calculateAmount(msg.Price, msg.Quantity, msg.PricePrecision)
		if err != nil {
			return "", -1, types.ErrInvalidOrderAmount("The frozen fee is too large")
		}
		amount = tmpAmount.RoundInt64()
	}
	if amount > types.MaxOrderAmount {
		return "", -1, types.ErrInvalidOrderAmount("The frozen fee is too large")
	}

	return denom, amount, nil
}

func handleMsgCreateOrder(ctx sdk.Context, msg types.MsgCreateOrder, keeper keepers.Keeper) sdk.Result {

	denom, amount, err := getDenomAndOrderAmount(msg)
	if err != nil {
		return err.Result()
	}
	seq, err := keeper.QuerySeqWithAddr(ctx, msg.Sender)
	if err != nil {
		return err.Result()
	}
	marketParams := keeper.GetParams(ctx)
	frozenFee, err := calOrderCommission(ctx, keeper, msg)
	if err != nil {
		return err.Result()
	}
	featureFee := calFeatureFeeForExistBlocks(msg, marketParams)
	totalFee := frozenFee + featureFee
	if featureFee > types.MaxOrderAmount || frozenFee > types.MaxOrderAmount || totalFee > types.MaxOrderAmount {
		return types.ErrInvalidOrderAmount("The frozen fee is too large").Result()
	}
	if err := checkMsgCreateOrder(ctx, keeper, msg, totalFee, amount, denom, seq); err != nil {
		return err.Result()
	}
	existBlocks := msg.ExistBlocks
	if existBlocks == 0 && msg.TimeInForce == GTE {
		existBlocks = marketParams.GTEOrderLifetime
	}

	order := types.Order{
		Sender:           msg.Sender,
		Sequence:         seq,
		Identify:         msg.Identify,
		TradingPair:      msg.TradingPair,
		OrderType:        msg.OrderType,
		Price:            sdk.NewDec(msg.Price).Quo(sdk.NewDec(int64(math.Pow10(int(msg.PricePrecision))))),
		Quantity:         msg.Quantity,
		Side:             msg.Side,
		TimeInForce:      msg.TimeInForce,
		Height:           ctx.BlockHeight(),
		ExistBlocks:      existBlocks,
		FrozenCommission: frozenFee,
		FrozenFeatureFee: featureFee,
		LeftStock:        msg.Quantity,
		Freeze:           amount,
		DealMoney:        0,
		DealStock:        0,
	}

	ork := keepers.NewOrderKeeper(keeper.GetMarketKey(), order.TradingPair, types.ModuleCdc)
	if err := ork.Add(ctx, &order); err != nil {
		return err.Result()
	}
	if err := handleFeeForCreateOrder(ctx, keeper, amount, denom, order.Sender, frozenFee, featureFee); err != nil {
		return err.Result()
	}
	sendCreateOrderMsg(ctx, keeper, order)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeKeyCreateOrder,
			sdk.NewAttribute(AttributeKeyOrder, order.OrderID()),
			sdk.NewAttribute(AttributeKeyTradingPair, order.TradingPair),
			sdk.NewAttribute(AttributeKeyHeight, strconv.FormatInt(order.Height, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	})
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

func checkMsgCreateOrder(ctx sdk.Context, keeper keepers.Keeper, msg types.MsgCreateOrder, cetFee int64, amount int64, denom string, seq uint64) sdk.Error {
	if cetFee != 0 {
		if !keeper.HasCoins(ctx, msg.Sender, sdk.Coins{sdk.NewCoin(dex.CET, sdk.NewInt(cetFee))}) {
			return types.ErrInsufficientCoins()
		}
	}
	stock, money := SplitSymbol(msg.TradingPair)
	totalAmount := sdk.NewInt(amount)
	if (stock == dex.CET && msg.Side == types.SELL) ||
		(money == dex.CET && msg.Side == types.BUY) {
		totalAmount = totalAmount.AddRaw(cetFee)
	}
	if !keeper.HasCoins(ctx, msg.Sender, sdk.Coins{sdk.NewCoin(denom, totalAmount)}) {
		return types.ErrInsufficientCoins()
	}
	orderID := types.AssemblyOrderID(msg.Sender.String(), seq, msg.Identify)
	globalKeeper := keepers.NewGlobalOrderKeeper(keeper.GetMarketKey(), types.ModuleCdc)
	if globalKeeper.QueryOrder(ctx, orderID) != nil {
		return types.ErrOrderAlreadyExist(orderID)
	}
	marketInfo, err := keeper.GetMarketInfo(ctx, msg.TradingPair)
	if err != nil {
		return types.ErrInvalidMarket(err.Error())
	}
	if p := msg.PricePrecision; p > marketInfo.PricePrecision {
		return types.ErrInvalidPricePrecision(p)
	}
	if keeper.IsTokenForbidden(ctx, stock) || keeper.IsTokenForbidden(ctx, money) {
		return types.ErrTokenForbidByIssuer()
	}
	if keeper.IsForbiddenByTokenIssuer(ctx, stock, msg.Sender) || keeper.IsForbiddenByTokenIssuer(ctx, money, msg.Sender) {
		return types.ErrAddressForbidByIssuer()
	}
	baseValue := types.GetGranularityOfOrder(marketInfo.OrderPrecision)
	if msg.Quantity%baseValue != 0 {
		return types.ErrInvalidOrderAmount("The amount of tokens to trade should be a multiple of the order precision")
	}

	return nil
}

func handleMsgCancelOrder(ctx sdk.Context, msg types.MsgCancelOrder, keeper keepers.Keeper) sdk.Result {
	if err := checkMsgCancelOrder(ctx, msg, keeper); err != nil {
		return err.Result()
	}
	order, marketParams := DoCancelOrder(ctx, keeper, msg.OrderID)

	// send msg to kafka
	sendCancelOrderMsg(ctx, order, &marketParams, keeper)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeKeyCancelOrder,
			sdk.NewAttribute(AttributeKeyOrder, order.OrderID()),
			sdk.NewAttribute(AttributeKeyDelOrderReason, types.CancelOrderByManual),
			sdk.NewAttribute(AttributeKeyDelOrderHeight, strconv.Itoa(int(ctx.BlockHeight()))),
			sdk.NewAttribute(AttributeKeyTradingPair, order.TradingPair),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	})
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
func DoCancelOrder(ctx sdk.Context, keeper keepers.Keeper, orderID string) (*types.Order, types.Params) {
	marketParams := keeper.GetParams(ctx)
	bankxKeeper := keeper.GetBankxKeeper()
	glk := keepers.NewGlobalOrderKeeper(keeper.GetMarketKey(), types.ModuleCdc)
	order := glk.QueryOrder(ctx, orderID)
	ork := keepers.NewOrderKeeper(keeper.GetMarketKey(), order.TradingPair, types.ModuleCdc)
	removeOrder(ctx, ork, bankxKeeper, keeper, order, &marketParams)
	return order, marketParams
}

func sendCancelOrderMsg(ctx sdk.Context, order *types.Order, params *Params, keeper keepers.Keeper) {
	if keeper.IsSubScribed(types.Topic) {
		cancelOrderInfo := packageCancelOrderMsgWithDelReason(ctx, order, types.CancelOrderByManual, params, keeper)
		msgqueue.FillMsgs(ctx, types.CancelOrderInfoKey, cancelOrderInfo)
	}
}

func checkMsgCancelOrder(ctx sdk.Context, msg types.MsgCancelOrder, keeper keepers.Keeper) sdk.Error {
	globalKeeper := keepers.NewGlobalOrderKeeper(keeper.GetMarketKey(), types.ModuleCdc)
	order := globalKeeper.QueryOrder(ctx, msg.OrderID)
	if order == nil {
		return types.ErrOrderNotFound(msg.OrderID)
	}
	if !bytes.Equal(order.Sender, msg.Sender) {
		return types.ErrNotMatchSender("only order's sender can cancel this order")
	}
	return nil
}

func handleMsgCancelTradingPair(ctx sdk.Context, msg types.MsgCancelTradingPair, keeper keepers.Keeper) sdk.Result {
	if err := checkMsgCancelTradingPair(keeper, msg, ctx); err != nil {
		return err.Result()
	}

	// Add del request to store
	dlk := keepers.NewDelistKeeper(keeper.GetMarketKey())
	if dlk.HasDelistRequest(ctx, msg.TradingPair) {
		return types.ErrDelistRequestExist(msg.TradingPair).Result()
	}
	dlk.AddDelistRequest(ctx, msg.EffectiveTime, msg.TradingPair)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeKeyCancelTradingPair,
			sdk.NewAttribute(AttributeKeyTradingPair, msg.TradingPair),
			sdk.NewAttribute(AttributeKeyEffectiveTime, strconv.Itoa(int(msg.EffectiveTime))),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

func checkMsgCancelTradingPair(keeper keepers.Keeper, msg types.MsgCancelTradingPair, ctx sdk.Context) sdk.Error {
	marketParams := keeper.GetParams(ctx)
	currTime := ctx.BlockHeader().Time.UnixNano()
	if msg.EffectiveTime < currTime+marketParams.MarketMinExpiredTime {
		return types.ErrInvalidCancelTime()
	}

	info, err := keeper.GetMarketInfo(ctx, msg.TradingPair)
	if err != nil {
		return types.ErrInvalidMarket(err.Error())
	}

	stockToken := keeper.GetToken(ctx, info.Stock)
	if !bytes.Equal(msg.Sender, stockToken.GetOwner()) {
		return types.ErrNotMatchSender("only stock's owner can cancel a market")
	}

	return nil
}

func calculateAmount(price, quantity int64, pricePrecision byte) (sdk.Dec, error) {
	actualPrice := sdk.NewDec(price).Quo(sdk.NewDec(int64(math.Pow10(int(pricePrecision)))))
	money := actualPrice.Mul(sdk.NewDec(quantity)).Add(sdk.NewDec(types.ExtraFrozenMoney)).Ceil()
	if money.GT(sdk.NewDec(types.MaxOrderAmount)) {
		return money, fmt.Errorf("exchange amount exceeds max int64 ")
	}
	return money, nil
}

func handleMsgModifyPricePrecision(ctx sdk.Context, msg types.MsgModifyPricePrecision, k keepers.Keeper) sdk.Result {
	if err := checkMsgModifyPricePrecision(ctx, msg, k); err != nil {
		return err.Result()
	}

	oldInfo, _ := k.GetMarketInfo(ctx, msg.TradingPair)
	info := types.MarketInfo{
		Stock:             oldInfo.Stock,
		Money:             oldInfo.Money,
		PricePrecision:    msg.PricePrecision,
		LastExecutedPrice: oldInfo.LastExecutedPrice,
	}
	if err := k.SetMarket(ctx, info); err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeKeyModifyPricePrecision,
			sdk.NewAttribute(AttributeKeyTradingPair, msg.TradingPair),
			sdk.NewAttribute(AttributeKeyOldPricePrecision, strconv.Itoa(int(oldInfo.PricePrecision))),
			sdk.NewAttribute(AttributeKeyNewPricePrecision, strconv.Itoa(int(info.PricePrecision))),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

func checkMsgModifyPricePrecision(ctx sdk.Context, msg types.MsgModifyPricePrecision, k keepers.Keeper) sdk.Error {
	_, err := k.GetMarketInfo(ctx, msg.TradingPair)
	if err != nil {
		return types.ErrInvalidMarket("Error retrieving market information: " + err.Error())
	}

	stock, _ := SplitSymbol(msg.TradingPair)
	tokenInfo := k.GetToken(ctx, stock)
	if !tokenInfo.GetOwner().Equals(msg.Sender) {
		return types.ErrNotMatchSender(fmt.Sprintf(
			"The sender of the transaction (%s) does not match the owner of the transaction pair (%s)",
			tokenInfo.GetOwner().String(), msg.Sender.String()))
	}

	return nil
}
