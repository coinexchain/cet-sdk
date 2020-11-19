package keepers

import (
	"fmt"
	"math"
	"math/big"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cet-sdk/msgqueue"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
)

var _ IPairKeeper = &PairKeeper{}

type Pair struct {
	Symbol string
}

type IPairKeeper interface {
	IPoolKeeper
	AddLimitOrder(ctx sdk.Context, order *types.Order) sdk.Error
	DeleteOrder(ctx sdk.Context, order types.MsgCancelOrder) sdk.Error
	HasOrder(ctx sdk.Context, orderID string) bool
	GetOrder(ctx sdk.Context, orderID string) *types.Order
	GetAllOrders(ctx sdk.Context, market string) []*types.Order

	SetParams(ctx sdk.Context, params types.Params)
	GetParams(ctx sdk.Context) types.Params
	GetTakerFee(ctx sdk.Context) sdk.Dec
	GetMakerFee(ctx sdk.Context) sdk.Dec
	GetDealWithPoolFee(ctx sdk.Context) sdk.Dec
	GetFeeToValidator(ctx sdk.Context) sdk.Dec

	ResetOrderIndexInOneBlock()
}

type FeeFunc func(sdk.Context) sdk.Dec

type PairKeeper struct {
	IPoolKeeper
	IOrderBookKeeper
	types.ExpectedAccountKeeper
	types.SupplyKeeper
	types.ExpectedBankKeeper
	codec       *codec.Codec
	storeKey    sdk.StoreKey
	subspace    params.Subspace
	msgProducer msgqueue.MsgSender
}

func NewPairKeeper(poolKeeper IPoolKeeper, supplyK types.SupplyKeeper, bnk types.ExpectedBankKeeper,
	accK types.ExpectedAccountKeeper, codec *codec.Codec, storeKey sdk.StoreKey, paramSubspace params.Subspace) *PairKeeper {
	return &PairKeeper{
		codec:                 codec,
		storeKey:              storeKey,
		IPoolKeeper:           poolKeeper,
		SupplyKeeper:          supplyK,
		ExpectedBankKeeper:    bnk,
		ExpectedAccountKeeper: accK,
		subspace:              paramSubspace.WithKeyTable(types.ParamKeyTable()),
		IOrderBookKeeper: &OrderKeeper{
			codec:    codec,
			storeKey: storeKey,
		},
	}
}

func (pk *PairKeeper) SetParams(ctx sdk.Context, params types.Params) {
	pk.subspace.SetParamSet(ctx, &params)
}

func (pk *PairKeeper) GetParams(ctx sdk.Context) (param types.Params) {
	pk.subspace.GetParamSet(ctx, &param)
	return
}

func (pk *PairKeeper) GetTakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(pk.GetParams(ctx).TakerFeeRateRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetMakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(pk.GetParams(ctx).MakerFeeRateRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetDealWithPoolFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(pk.GetParams(ctx).DealWithPoolFeeRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetFeeToValidator(ctx sdk.Context) sdk.Dec {
	param := pk.GetParams(ctx)
	return sdk.NewDec(param.FeeToValidator).QuoInt64(param.FeeToValidator + param.FeeToPool)
}

func (pk *PairKeeper) AllocateFeeToValidatorAndPool(ctx sdk.Context, denom string, totalAmount sdk.Int, sender sdk.AccAddress) (sdk.Int, sdk.Error) {
	feeToVal := pk.GetFeeToValidator(ctx).MulInt(totalAmount).TruncateInt()
	feeToPool := totalAmount.Sub(feeToVal)
	err := pk.SendCoinsFromAccountToModule(ctx, sender, auth.FeeCollectorName, sdk.NewCoins(sdk.NewCoin(denom, feeToVal)))
	if err != nil {
		return sdk.ZeroInt(), err
	}
	err = pk.SendCoinsFromAccountToModule(ctx, sender, types.PoolModuleAcc, sdk.NewCoins(sdk.NewCoin(denom, feeToPool)))
	if err != nil {
		return sdk.ZeroInt(), err
	}

	return feeToPool, nil
}

func (pk *PairKeeper) AllocateFeeToValidator(ctx sdk.Context, fee sdk.Coins, sender sdk.AccAddress) sdk.Error {
	if err := pk.UnFreezeCoins(ctx, sender, fee); err != nil {
		panic(err)
	}
	err := pk.SendCoinsFromAccountToModule(ctx, sender, auth.FeeCollectorName, fee)
	if err != nil {
		return err
	}
	return nil
}

func (pk *PairKeeper) AddLimitOrder(ctx sdk.Context, order *types.Order) (err sdk.Error) {
	defer func() {
		r := recover()
		switch r.(type) {
		case sdk.Error:
			err = r.(sdk.Error)
		case string:
			err = sdk.NewError(types.RouterKey, types.CodeUnKnownError, r.(string))
		case error:
			err = sdk.NewError(types.RouterKey, types.CodeUnKnownError, r.(error).Error())
		}
	}()

	poolInfo := pk.GetPoolInfo(ctx, order.TradingPair)
	if poolInfo == nil {
		return types.ErrInvalidMarket(order.TradingPair)
	}
	if order.PricePrecision > poolInfo.PricePrecision {
		return types.ErrInvalidPricePrecision(order.PricePrecision, poolInfo.PricePrecision)
	}
	order.Sequence = int64(pk.GetAccount(ctx, order.Sender).GetSequence())
	if pk.GetOrder(ctx, order.GetOrderID()) != nil {
		return types.ErrOrderAlreadyExist(order.GetOrderID())
	}
	actualAmount, err := pk.freezeOrderCoin(ctx, order)
	if err != nil {
		return err
	}
	dealInfo := &types.DealInfo{
		RemainAmount:    actualAmount,
		AmountInToPool:  sdk.ZeroInt(),
		DealMoneyInBook: sdk.ZeroInt(),
		DealStockInBook: sdk.ZeroInt(),
	}

	// 1. get matched opposite orders.
	oppositeOrders := pk.GetMatchedOrder(ctx, order)
	// 2. deal order with pool
	if len(oppositeOrders) > 0 {
		for _, opOrder := range oppositeOrders {
			if allDeal, err := pk.dealOrderWithOrderBookAndPool(ctx, order, opOrder, dealInfo, poolInfo); allDeal {
				break
			} else if err != nil {
				return err
			}
		}
	}
	pk.tryDealInPool(dealInfo, order.Price, order, poolInfo)

	// 3. update poolInfo with new order
	if dealInfo.AmountInToPool.IsPositive() {
		if order.IsBuy {
			poolInfo.MoneyOrderBookReserve = poolInfo.MoneyOrderBookReserve.Add(actualAmount.Sub(dealInfo.AmountInToPool))
		} else {
			poolInfo.StockOrderBookReserve = poolInfo.StockOrderBookReserve.Add(actualAmount.Sub(dealInfo.AmountInToPool))
		}
	} else {
		if order.IsBuy {
			poolInfo.MoneyOrderBookReserve = poolInfo.MoneyOrderBookReserve.Add(actualAmount)
		} else {
			poolInfo.StockOrderBookReserve = poolInfo.StockOrderBookReserve.Add(actualAmount)
		}
	}

	// 3. final deal with pool and order
	pk.finalDealWithPool(ctx, order, dealInfo, poolInfo)

	// 4. store order in keeper if need
	pk.storeOrderIfNeed(ctx, order, poolInfo)
	pk.SetPoolInfo(ctx, order.TradingPair, poolInfo)
	return nil
}

func (pk PairKeeper) sendCreateOrderInfo(ctx sdk.Context, order *types.Order) {
	if pk.msgProducer == nil || pk.msgProducer.IsSubscribed(types.ModuleName) {
		return
	}
	info := types.CreateOrderInfoMq{
		TradingPair: order.TradingPair,
		Height:      ctx.BlockHeight(),
	}
	msgqueue.FillMsgs(ctx, types.CreateOrderInfoKey, info)
}

func (pk PairKeeper) freezeOrderCoin(ctx sdk.Context, order *types.Order) (sdk.Int, sdk.Error) {
	orderAmount := order.ActualAmount()
	if order.IsBuy {
		if err := pk.FreezeCoins(ctx, order.Sender, newCoins(order.Money(), orderAmount)); err != nil {
			return sdk.Int{}, err
		}
	} else {
		if err := pk.FreezeCoins(ctx, order.Sender, newCoins(order.Stock(), orderAmount)); err != nil {
			return sdk.Int{}, err
		}
	}
	return orderAmount, nil
}

func (pk PairKeeper) dealOrderWithOrderBookAndPool(ctx sdk.Context, order, oppositeOrder *types.Order,
	dealInfo *types.DealInfo, poolInfo *PoolInfo) (allDeal bool, err sdk.Error) {
	pk.tryDealInPool(dealInfo, oppositeOrder.Price, order, poolInfo)
	pk.dealInOrderBook(ctx, order, oppositeOrder, poolInfo, dealInfo, !poolInfo.IsNoReservePool())
	if oppositeOrder.LeftStock == 0 {
		pk.DelOrder(ctx, oppositeOrder)
	} else {
		pk.StoreToOrderBook(ctx, oppositeOrder)
	}
	return order.LeftStock == 0, nil
}

func (pk PairKeeper) tryDealInPool(dealInfo *types.DealInfo, dealPrice sdk.Dec, order *types.Order, info *PoolInfo) bool {
	currTokenCanTradeWithPool := IntoPoolAmountTillPrice(dealPrice, order.IsBuy, info)
	if currTokenCanTradeWithPool.GT(dealInfo.AmountInToPool) {
		diffTokenTradeWithPool := currTokenCanTradeWithPool.Sub(dealInfo.AmountInToPool)
		allDeal := diffTokenTradeWithPool.GT(dealInfo.RemainAmount)
		if allDeal {
			diffTokenTradeWithPool = dealInfo.RemainAmount
		}
		before := GetAmountOutInPool(dealInfo.AmountInToPool, info, order.IsBuy)
		after := GetAmountOutInPool(currTokenCanTradeWithPool, info, order.IsBuy)
		order.Freeze -= diffTokenTradeWithPool.Int64()
		if order.IsBuy {
			order.LeftStock -= after.Sub(before).Int64()
			order.DealStock += after.Sub(before).Int64()
			order.DealMoney += currTokenCanTradeWithPool.Sub(dealInfo.AmountInToPool).Int64()
		} else {
			order.LeftStock -= currTokenCanTradeWithPool.Sub(dealInfo.AmountInToPool).Int64()
			order.DealStock += currTokenCanTradeWithPool.Sub(dealInfo.AmountInToPool).Int64()
			order.DealMoney += after.Sub(before).Int64()
		}
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(diffTokenTradeWithPool)
		dealInfo.AmountInToPool = dealInfo.AmountInToPool.Add(diffTokenTradeWithPool)
		return allDeal
	}
	return false
}

func IntoPoolAmountTillPrice(dealPrice sdk.Dec, isBuy bool, info *PoolInfo) sdk.Int {
	if isBuy {
		root := dealPrice.Mul(sdk.NewDecFromInt(info.StockAmmReserve)).Mul(sdk.NewDecFromInt(info.MoneyAmmReserve)).MulInt64(int64(math.Pow10(16)))
		//fmt.Println("root: ", root, "; sqrt(root): ", big.NewInt(0).Sqrt(root.TruncateInt().BigInt()))
		//if ret := sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.TruncateInt().
		//	BigInt())).Quo(sdk.NewDec(int64(math.Pow10(8)))).Sub(sdk.NewDecFromInt(info.MoneyAmmReserve)).TruncateInt(); ret.IsPositive() {
		root = sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.TruncateInt().BigInt()))
		if root.LTE(sdk.NewDec(int64(math.Pow10(8)))) {
			return sdk.ZeroInt()
		}
		if ret := root.Quo(sdk.NewDec(int64(math.Pow10(8)))).Sub(sdk.NewDecFromInt(info.MoneyAmmReserve)).TruncateInt(); ret.IsPositive() {
			return ret
		}
		return sdk.ZeroInt()
	}
	root := sdk.NewDecFromInt(info.MoneyAmmReserve).Mul(sdk.NewDecFromInt(info.StockAmmReserve)).MulInt64(int64(math.Pow10(16))).Quo(dealPrice)
	fmt.Println("root: ", root, "; sqrt(root): ", big.NewInt(0).Sqrt(root.TruncateInt().BigInt()))
	root = sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.TruncateInt().BigInt()))
	if root.LTE(sdk.NewDec(int64(math.Pow10(8)))) {
		return sdk.ZeroInt()
	}
	if ret := root.Quo(sdk.NewDec(int64(math.Pow10(8)))).Sub(sdk.NewDecFromInt(info.StockAmmReserve)).TruncateInt(); ret.IsPositive() {
		return ret
	}
	return sdk.ZeroInt()
}

func GetAmountOutInPool(amountIn sdk.Int, poolInfo *PoolInfo, isBuy bool) sdk.Int {
	outPoolTokenReserve, inPoolTokenReserve := poolInfo.MoneyAmmReserve, poolInfo.StockAmmReserve
	if isBuy {
		outPoolTokenReserve, inPoolTokenReserve = poolInfo.StockAmmReserve, poolInfo.MoneyAmmReserve
	}
	return outPoolTokenReserve.Mul(amountIn).Quo(inPoolTokenReserve.Add(amountIn))
}

func (pk PairKeeper) dealInOrderBook(ctx sdk.Context, currOrder,
	orderInBook *types.Order, poolInfo *PoolInfo, dealInfo *types.DealInfo, isPoolExists bool) {
	if currOrder.LeftStock == 0 {
		return
	}
	// calculate stock amount
	dealStockAmount := currOrder.LeftStock
	if orderInBook.LeftStock < currOrder.LeftStock {
		dealStockAmount = orderInBook.LeftStock
	}

	var (
		stockFee   sdk.Int
		moneyFee   sdk.Int
		takerFee   sdk.Int
		makerFee   sdk.Int
		stockTrans = sdk.NewInt(dealStockAmount)
		moneyTrans = sdk.NewDec(dealStockAmount).Mul(orderInBook.Price).TruncateInt() // less
	)
	currOrder.LeftStock -= dealStockAmount
	orderInBook.LeftStock -= dealStockAmount
	currOrder.DealMoney += moneyTrans.Int64()
	currOrder.DealStock += stockTrans.Int64()
	orderInBook.DealStock += stockTrans.Int64()
	orderInBook.DealMoney += moneyTrans.Int64()
	poolInfo.StockOrderBookReserve = poolInfo.StockOrderBookReserve.Sub(stockTrans)
	poolInfo.MoneyOrderBookReserve = poolInfo.MoneyOrderBookReserve.Sub(moneyTrans)
	makeFeeRate := sdk.NewDec(pk.GetParams(ctx).MakerFeeRateRate)
	takerFeeRate := sdk.NewDec(pk.GetParams(ctx).TakerFeeRateRate)
	if currOrder.IsBuy {
		currOrder.Freeze -= moneyTrans.Int64()
		orderInBook.Freeze -= stockTrans.Int64()
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(moneyTrans)
		moneyFee = makeFeeRate.MulInt(moneyTrans).Add(
			sdk.NewDec(9999)).Quo(sdk.NewDec(types.DefaultFeePrecision)).TruncateInt()
		stockFee = takerFeeRate.MulInt(stockTrans).Add(
			sdk.NewDec(9999)).Quo(sdk.NewDec(types.DefaultFeePrecision)).TruncateInt()
	} else {
		currOrder.Freeze -= stockTrans.Int64()
		orderInBook.Freeze -= moneyTrans.Int64()
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(stockTrans)
		stockFee = makeFeeRate.MulInt(stockTrans).Add(
			sdk.NewDec(9999)).Quo(sdk.NewDec(types.DefaultFeePrecision)).TruncateInt()
		moneyFee = takerFeeRate.MulInt(moneyTrans).Add(
			sdk.NewDec(9999)).Quo(sdk.NewDec(types.DefaultFeePrecision)).TruncateInt()
	}

	dealInfo.DealMoneyInBook = dealInfo.DealMoneyInBook.Add(moneyTrans)
	dealInfo.DealStockInBook = dealInfo.DealStockInBook.Add(stockTrans)
	// transfer tokens in orders
	if currOrder.IsBuy {
		takerFee = stockFee
		makerFee = moneyFee
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Money(), moneyTrans.Sub(moneyFee))
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Stock(), stockTrans.Sub(stockFee))
		if isPoolExists {
			moneyToPool, err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Money(), moneyFee, currOrder.Sender)
			if err != nil {
				panic(err)
			}
			stockToPool, err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Stock(), stockFee, orderInBook.Sender)
			if err != nil {
				panic(err)
			}
			poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Add(moneyToPool)
			poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Add(stockToPool)
		} else {
			if err := pk.AllocateFeeToValidator(ctx, sdk.NewCoins(sdk.NewCoin(currOrder.Money(), moneyFee)), currOrder.Sender); err != nil {
				panic(err)
			}
			if err := pk.AllocateFeeToValidator(ctx, sdk.NewCoins(sdk.NewCoin(currOrder.Stock(), stockFee)), orderInBook.Sender); err != nil {
				panic(err)
			}
		}
	} else {
		takerFee = moneyFee
		makerFee = stockFee
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Stock(), stockTrans.Sub(stockFee))
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Money(), moneyTrans.Sub(moneyFee))
		if isPoolExists {
			stockToPool, err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Stock(), stockFee, currOrder.Sender)
			if err != nil {
				panic(err)
			}
			moneyToPool, err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Money(), moneyFee, orderInBook.Sender)
			if err != nil {
				panic(err)
			}
			poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Add(moneyToPool)
			poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Add(stockToPool)
		} else {
			if err := pk.AllocateFeeToValidator(ctx, sdk.NewCoins(sdk.NewCoin(currOrder.Stock(), stockFee)), currOrder.Sender); err != nil {
				panic(err)
			}
			if err := pk.AllocateFeeToValidator(ctx, sdk.NewCoins(sdk.NewCoin(currOrder.Money(), moneyFee)), orderInBook.Sender); err != nil {
				panic(err)
			}
		}
	}
	pk.sendDealOrderMsg(ctx, currOrder, orderInBook, dealStockAmount, moneyTrans.Int64(), takerFee.Int64(), makerFee.Int64(), poolInfo)
}

func (pk PairKeeper) sendDealOrderMsg(ctx sdk.Context, order, dealOrder *types.Order, dealStock, dealMoney int64, takerFee, makerFee int64, poolInfo *PoolInfo) {
	if pk.msgProducer == nil || pk.msgProducer.IsSubscribed(types.ModuleName) {
		return
	}
	taker := types.FillOrderInfoMq{
		OrderID:     order.GetOrderID(),
		TradingPair: order.TradingPair,
		Height:      ctx.BlockHeight(),
		Price:       order.Price,
		Side:        getSide(order.IsBuy),

		LeftStock:          order.LeftStock,
		Freeze:             order.Freeze,
		DealStock:          order.DealStock,
		DealMoney:          order.DealMoney,
		CurrMoney:          dealMoney,
		CurrStock:          dealStock,
		FillPrice:          dealOrder.Price,
		CurrUsedCommission: takerFee,
	}
	maker := types.FillOrderInfoMq{
		OrderID:     dealOrder.GetOrderID(),
		TradingPair: dealOrder.TradingPair,
		Height:      ctx.BlockHeight(),
		Price:       dealOrder.Price,
		Side:        getSide(dealOrder.IsBuy),

		LeftStock:          dealOrder.LeftStock,
		Freeze:             dealOrder.Freeze,
		DealStock:          dealOrder.DealStock,
		DealMoney:          dealOrder.DealMoney,
		CurrMoney:          dealMoney,
		CurrStock:          dealStock,
		FillPrice:          dealOrder.Price,
		CurrUsedCommission: makerFee,
	}
	dealMarketInfo := types.MarketDealInfoMq{
		TradingPair:     order.TradingPair,
		TakerOrderID:    order.GetOrderID(),
		MakerOrderID:    dealOrder.GetOrderID(),
		DealStockAmount: dealStock,
		DealHeight:      ctx.BlockHeight(),
		DealPrice:       dealOrder.Price,
	}
	msgqueue.FillMsgs(ctx, types.FillOrderInfoKey, taker)
	msgqueue.FillMsgs(ctx, types.FillOrderInfoKey, maker)
	msgqueue.FillMsgs(ctx, types.DealMarketInfoKey, dealMarketInfo)

	if order.LeftStock == 0 {
		pk.sendDelOrderInfo(ctx, order, types.CancelOrderByAllFilled)
	}
	if dealOrder.LeftStock == 0 {
		pk.sendDelOrderInfo(ctx, dealOrder, types.CancelOrderByAllFilled)
	}

	// update last execute price in pool
	poolInfo.LastExecutedPrice = dealOrder.Price
}

func getSide(isBuy bool) byte {
	if isBuy {
		return types.BID
	}
	return types.ASK
}

func (pk PairKeeper) sendDelOrderInfo(ctx sdk.Context, order *types.Order, delReason string) {
	if pk.msgProducer == nil || pk.msgProducer.IsSubscribed(types.ModuleName) {
		return
	}
	info := types.CancelOrderInfoMq{
		OrderID:     order.GetOrderID(),
		TradingPair: order.TradingPair,
		Height:      ctx.BlockHeight(),
		Side:        getSide(order.IsBuy),
		Price:       order.Price,

		DelReason:    delReason,
		LeftStock:    order.LeftStock,
		RemainAmount: order.Freeze,
		DealStock:    order.DealStock,
		DealMoney:    order.DealMoney,
	}
	msgqueue.FillMsgs(ctx, types.CancelOrderInfoKey, info)
}

func (pk PairKeeper) finalDealWithPool(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) {
	_, fee, poolToUser := pk.dealWithPoolAndCollectFee(ctx, order, dealInfo, poolInfo)
	if dealInfo.AmountInToPool.IsPositive() {
		pk.sendDealInfoWithPool(ctx, dealInfo, order, fee.Int64(), poolToUser.Int64(), poolInfo)
	}
}

func (pk PairKeeper) dealWithPoolAndCollectFee(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) (totalAmountToTaker sdk.Int, fee sdk.Int, poolToUser sdk.Int) {
	otherToTaker := dealInfo.DealMoneyInBook
	if order.IsBuy {
		otherToTaker = dealInfo.DealStockInBook
	}
	outAmount := sdk.ZeroInt()
	if dealInfo.AmountInToPool.IsPositive() {
		outAmount = GetAmountOutInPool(dealInfo.AmountInToPool, poolInfo, order.IsBuy)
		//outAmount = outPoolTokenReserve.Mul(dealInfo.AmountInToPool).Quo(inPoolTokenReserve.Add(dealInfo.AmountInToPool))
	} else {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	}

	// add fee calculate
	fee = pk.GetDealWithPoolFee(ctx).MulInt(outAmount).TruncateInt()
	amountToTaker := outAmount.Add(otherToTaker).Sub(fee)
	if order.IsBuy {
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Sub(outAmount)
	} else {
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Sub(outAmount)
	}
	if order.IsBuy {
		if err := pk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, order.Sender, newCoins(order.Stock(), outAmount)); err != nil {
			panic(err)
		}
		stockToPool, err := pk.AllocateFeeToValidatorAndPool(ctx, order.Stock(), fee, order.Sender)
		if err != nil {
			panic(err)
		}
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Add(stockToPool)
		if err := pk.SendCoinsFromAccountToModule(ctx, order.Sender, types.PoolModuleAcc, newCoins(order.Money(), dealInfo.AmountInToPool)); err != nil {
			panic(err)
		}
	} else {
		if err := pk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, order.Sender, newCoins(order.Money(), outAmount)); err != nil {
			panic(err)
		}
		moneyToPool, err := pk.AllocateFeeToValidatorAndPool(ctx, order.Money(), fee, order.Sender)
		if err != nil {
			panic(err)
		}
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Add(moneyToPool)
		if err := pk.SendCoinsFromAccountToModule(ctx, order.Sender, types.PoolModuleAcc, newCoins(order.Stock(), dealInfo.AmountInToPool)); err != nil {
			panic(err)
		}
	}
	return amountToTaker, fee, outAmount
}

func (pk PairKeeper) sendDealInfoWithPool(ctx sdk.Context, dealInfo *types.DealInfo, order *types.Order, commission, poolAmount int64, poolInfo *PoolInfo) {
	if pk.msgProducer == nil || pk.msgProducer.IsSubscribed(types.ModuleName) {
		return
	}
	currStock := poolAmount
	currMoney := dealInfo.AmountInToPool.Int64()
	if !order.IsBuy {
		currStock = dealInfo.AmountInToPool.Int64()
		currMoney = poolAmount
	}
	dealOrderInfo := types.FillOrderInfoMq{
		OrderID:            order.GetOrderID(),
		TradingPair:        order.TradingPair,
		Height:             ctx.BlockHeight(),
		Side:               getSide(order.IsBuy),
		Price:              order.Price,
		Freeze:             order.Freeze,
		LeftStock:          order.LeftStock,
		DealMoney:          order.DealMoney,
		DealStock:          order.DealStock,
		CurrStock:          currStock,
		CurrMoney:          currMoney,
		FillPrice:          sdk.NewDec(currMoney).Quo(sdk.NewDec(currStock)),
		CurrUsedCommission: commission,
	}
	dealMarketInfo := types.MarketDealInfoMq{
		TradingPair:     order.TradingPair,
		MakerOrderID:    order.GetOrderID(),
		TakerOrderID:    types.ReservePoolID,
		DealStockAmount: currStock,
		DealHeight:      ctx.BlockHeight(),
		DealPrice:       dealOrderInfo.FillPrice,
	}
	msgqueue.FillMsgs(ctx, types.FillOrderInfoKey, dealOrderInfo)
	msgqueue.FillMsgs(ctx, types.DealMarketInfoKey, dealMarketInfo)
	// update pool last execute price
	poolInfo.LastExecutedPrice = dealOrderInfo.FillPrice
}

func (pk PairKeeper) storeOrderIfNeed(ctx sdk.Context, order *types.Order, poolInfo *PoolInfo) {
	if order.LeftStock > 0 {
		pk.AddOrder(ctx, order)
	}
}

func (pk PairKeeper) transferToken(ctx sdk.Context, from, to sdk.AccAddress, token string, amount sdk.Int) {
	coin := newCoins(token, amount)
	if err := pk.UnFreezeCoins(ctx, from, coin); err != nil {
		panic(err)
	}
	if err := pk.SendCoins(ctx, from, to, coin); err != nil {
		panic(err)
	}
}

func newCoins(token string, amount sdk.Int) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(token, amount))
}

func (pk PairKeeper) HasOrder(ctx sdk.Context, orderID string) bool {
	return pk.GetOrder(ctx, orderID) != nil
}

func (pk PairKeeper) GetOrder(ctx sdk.Context, orderID string) *types.Order {
	return pk.IOrderBookKeeper.GetOrder(ctx, &QueryOrderInfo{
		OrderID: orderID,
	})
}

func (pk *PairKeeper) DeleteOrder(ctx sdk.Context, msg types.MsgCancelOrder) sdk.Error {
	order := pk.GetOrder(ctx, msg.OrderID)
	if order == nil {
		return types.ErrInvalidOrderID(msg.OrderID)
	}
	if !order.Sender.Equals(msg.Sender) {
		return types.ErrInvalidSender(msg.Sender, order.Sender)
	}
	pk.IOrderBookKeeper.DelOrder(ctx, order)
	if err := pk.updateOrderBookReserveByOrderDel(ctx, order); err != nil {
		return err
	}
	pk.sendDelOrderInfo(ctx, order, types.CancelOrderByManual)
	return nil
}

func (pk PairKeeper) updateOrderBookReserveByOrderDel(ctx sdk.Context, delOrder *types.Order) sdk.Error {
	info := pk.GetPoolInfo(ctx, delOrder.TradingPair)
	if delOrder.IsBuy {
		amount := delOrder.ActualAmount()
		info.MoneyOrderBookReserve = info.MoneyOrderBookReserve.Sub(amount)
		if err := pk.UnFreezeCoins(ctx, delOrder.Sender, newCoins(delOrder.Money(), amount)); err != nil {
			return err
		}
	} else {
		info.StockOrderBookReserve = info.StockOrderBookReserve.Sub(sdk.NewInt(delOrder.LeftStock))
		if err := pk.UnFreezeCoins(ctx, delOrder.Sender, newCoins(delOrder.Stock(), sdk.NewInt(delOrder.LeftStock))); err != nil {
			return err
		}
	}
	pk.SetPoolInfo(ctx, delOrder.TradingPair, info)
	return nil
}

func (pk *PairKeeper) GetAllOrders(ctx sdk.Context, market string) []*types.Order {
	return pk.GetAllOrdersInMarket(ctx, market)
}
