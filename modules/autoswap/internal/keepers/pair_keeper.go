package keepers

import (
	"math"

	"github.com/coinexchain/cet-sdk/modules/market"

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
	GetOrdersFromUser(ctx sdk.Context, user string) []string

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
	types.ExpectedAuthXKeeper
	types.ExpectedAccountKeeper
	types.SupplyKeeper
	types.ExpectedBankKeeper
	codec       *codec.Codec
	storeKey    sdk.StoreKey
	subspace    params.Subspace
	msgProducer msgqueue.MsgSender
}

func NewPairKeeper(poolKeeper IPoolKeeper, supplyK types.SupplyKeeper, bnk types.ExpectedBankKeeper,
	accK types.ExpectedAccountKeeper, accxK types.ExpectedAuthXKeeper, codec *codec.Codec, storeKey sdk.StoreKey, paramSubspace params.Subspace) *PairKeeper {
	return &PairKeeper{
		codec:                 codec,
		storeKey:              storeKey,
		IPoolKeeper:           poolKeeper,
		SupplyKeeper:          supplyK,
		ExpectedBankKeeper:    bnk,
		ExpectedAccountKeeper: accK,
		ExpectedAuthXKeeper:   accxK,
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
	return sdk.NewDec(pk.GetParams(ctx).TakerFeeRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetMakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(pk.GetParams(ctx).MakerFeeRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetDealWithPoolFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(pk.GetParams(ctx).DealWithPoolFeeRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetFeeToValidator(ctx sdk.Context) sdk.Dec {
	param := pk.GetParams(ctx)
	return sdk.NewDec(param.FeeToValidator).QuoInt64(param.FeeToValidator + param.FeeToPool)
}

func (pk *PairKeeper) AllocateFeeToValidatorAndPool(ctx sdk.Context, denom string, totalAmount sdk.Int, sender sdk.AccAddress) (sdk.Int, sdk.Int, sdk.AccAddress, sdk.Error) {
	var (
		reference    sdk.AccAddress
		rebateAmount sdk.Int
	)
	rebateAmount, reference, err := pk.sendRebateAmount(ctx, denom, totalAmount, sender)
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), nil, err
	}
	totalAmount = totalAmount.Sub(rebateAmount)
	feeToVal := pk.GetFeeToValidator(ctx).MulInt(totalAmount).TruncateInt()
	feeToPool := totalAmount.Sub(feeToVal)
	err = pk.SendCoinsFromAccountToModule(ctx, sender, auth.FeeCollectorName, sdk.NewCoins(sdk.NewCoin(denom, feeToVal)))
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), nil, err
	}
	err = pk.SendCoinsFromAccountToModule(ctx, sender, types.PoolModuleAcc, sdk.NewCoins(sdk.NewCoin(denom, feeToPool)))
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), nil, err
	}
	return feeToPool, rebateAmount, reference, nil
}

func (pk *PairKeeper) AllocateFeeToValidator(ctx sdk.Context, denom string, fee sdk.Int, sender sdk.AccAddress) (sdk.Int, sdk.AccAddress, sdk.Error) {
	if err := pk.UnFreezeCoins(ctx, sender, newCoins(denom, fee)); err != nil {
		panic(err)
	}
	rebateAmount, reference, err := pk.sendRebateAmount(ctx, denom, fee, sender)
	if err != nil {
		return sdk.ZeroInt(), nil, err
	}
	fee = fee.Sub(rebateAmount)
	if err = pk.SendCoinsFromAccountToModule(ctx, sender, auth.FeeCollectorName, newCoins(denom, fee)); err != nil {
		return sdk.ZeroInt(), nil, err
	}
	return rebateAmount, reference, nil
}

func (pk PairKeeper) sendRebateAmount(ctx sdk.Context, denom string, totalFee sdk.Int, sender sdk.AccAddress) (sdk.Int, sdk.AccAddress, sdk.Error) {
	var (
		reference    sdk.AccAddress
		rebateAmount sdk.Int
	)
	if reference = pk.GetRefereeAddr(ctx, sender); reference == nil {
		return sdk.ZeroInt(), nil, nil
	}
	rebateAmount = pk.calRebateAmount(ctx, totalFee)
	if err := pk.SendCoins(ctx, sender, reference,
		sdk.NewCoins(sdk.NewCoin(denom, rebateAmount))); err != nil {
		return sdk.ZeroInt(), reference, err
	}
	return rebateAmount, reference, nil
}

func (pk PairKeeper) calRebateAmount(ctx sdk.Context, fee sdk.Int) sdk.Int {
	ratio := pk.GetRebateRatio(ctx)
	ratioBase := pk.GetRebateRatioBase(ctx)
	rebateAmount := fee.MulRaw(ratio).QuoRaw(ratioBase)
	return rebateAmount
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
	if tmp := pk.GetOrder(ctx, order.GetOrderID()); tmp != nil {
		return types.ErrOrderAlreadyExist(order.GetOrderID())
	}
	actualAmount, err := pk.freezeOrderCoin(ctx, order)
	if err != nil {
		return err
	}
	dealInfo := &types.DealInfo{
		RemainAmount:      actualAmount,
		AmountInToPool:    sdk.ZeroInt(),
		DealMoneyInBook:   sdk.ZeroInt(),
		DealStockInBook:   sdk.ZeroInt(),
		FeeToStockReserve: sdk.ZeroInt(),
		FeeToMoneyReserve: sdk.ZeroInt(),
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
		if oppositeOrder.Freeze > 0 {
			if oppositeOrder.IsBuy {
				if err := pk.UnFreezeCoins(ctx, oppositeOrder.Sender, newCoins(oppositeOrder.Money(), sdk.NewInt(oppositeOrder.Freeze))); err != nil {
					return false, err
				}
			} else {
				if err := pk.UnFreezeCoins(ctx, oppositeOrder.Sender, newCoins(oppositeOrder.Stock(), sdk.NewInt(oppositeOrder.Freeze))); err != nil {
					return false, err
				}
			}
		}
	} else {
		pk.StoreToOrderBook(ctx, oppositeOrder)
	}
	return order.LeftStock == 0, nil
}

func (pk PairKeeper) tryDealInPool(dealInfo *types.DealInfo, dealPrice sdk.Dec, order *types.Order, info *PoolInfo) bool {
	currTokenCanTradeWithPool := IntoPoolAmountTillPrice(dealPrice, order.IsBuy, info)
	if currTokenCanTradeWithPool.GT(dealInfo.AmountInToPool) && dealInfo.RemainAmount.IsPositive() {
		diffTokenTradeWithPool := currTokenCanTradeWithPool.Sub(dealInfo.AmountInToPool)
		allDeal := diffTokenTradeWithPool.GT(dealInfo.RemainAmount)
		if allDeal {
			diffTokenTradeWithPool = dealInfo.RemainAmount
		}
		before := GetAmountOutInPool(dealInfo.AmountInToPool, info, order.IsBuy)
		after := GetAmountOutInPool(currTokenCanTradeWithPool, info, order.IsBuy)
		if after.Sub(before).IsZero() {
			return false
		}
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
		root := dealPrice.Mul(sdk.NewDecFromInt(info.StockAmmReserve)).Mul(sdk.NewDecFromInt(info.MoneyAmmReserve)).MulInt64(int64(math.Pow10(10)))
		root = sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.TruncateInt().BigInt()))
		if root.LTE(sdk.NewDec(int64(math.Pow10(5)))) {
			return sdk.ZeroInt()
		}
		if ret := root.Quo(sdk.NewDec(int64(math.Pow10(5)))).Sub(sdk.NewDecFromInt(info.MoneyAmmReserve)).TruncateInt(); ret.IsPositive() {
			return ret
		}
		return sdk.ZeroInt()
	}
	root := sdk.NewDecFromInt(info.MoneyAmmReserve).Mul(sdk.NewDecFromInt(info.StockAmmReserve)).MulInt64(int64(math.Pow10(10))).Quo(dealPrice)
	root = sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.TruncateInt().BigInt()))
	if root.LTE(sdk.NewDec(int64(math.Pow10(5)))) {
		return sdk.ZeroInt()
	}
	if ret := root.Quo(sdk.NewDec(int64(math.Pow10(5)))).Sub(sdk.NewDecFromInt(info.StockAmmReserve)).TruncateInt(); ret.IsPositive() {
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
		stockFee                sdk.Int
		moneyFee                sdk.Int
		takerFee                sdk.Int
		makerFee                sdk.Int
		moneyToPool             sdk.Int
		stockToPool             sdk.Int
		err                     sdk.Error
		stockTrans              = sdk.NewInt(dealStockAmount)
		moneyTrans              = sdk.NewDec(dealStockAmount).Mul(orderInBook.Price).TruncateInt() // less
		takerRebateAmount       sdk.Int
		makerRebateAmount       sdk.Int
		takerOrderReferenceAddr sdk.AccAddress
		makerOrderReferenceAddr sdk.AccAddress
	)
	if currOrder.IsBuy {
		if moneyTrans.GT(dealInfo.RemainAmount) {
			moneyTrans = dealInfo.RemainAmount
		}
	} else {
		if stockTrans.GT(dealInfo.RemainAmount) {
			stockTrans = dealInfo.RemainAmount
		}
	}
	currOrder.LeftStock -= dealStockAmount
	orderInBook.LeftStock -= dealStockAmount
	currOrder.DealMoney += moneyTrans.Int64()
	currOrder.DealStock += stockTrans.Int64()
	orderInBook.DealStock += stockTrans.Int64()
	orderInBook.DealMoney += moneyTrans.Int64()
	poolInfo.StockOrderBookReserve = poolInfo.StockOrderBookReserve.Sub(stockTrans)
	poolInfo.MoneyOrderBookReserve = poolInfo.MoneyOrderBookReserve.Sub(moneyTrans)
	makeFeeRate := sdk.NewDec(pk.GetParams(ctx).MakerFeeRate)
	takerFeeRate := sdk.NewDec(pk.GetParams(ctx).TakerFeeRate)
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
			moneyToPool, takerRebateAmount, takerOrderReferenceAddr, err = pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Money(), moneyFee, currOrder.Sender)
			if err != nil {
				panic(err)
			}
			stockToPool, makerRebateAmount, makerOrderReferenceAddr, err = pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Stock(), stockFee, orderInBook.Sender)
			if err != nil {
				panic(err)
			}
			dealInfo.FeeToMoneyReserve = dealInfo.FeeToMoneyReserve.Add(moneyToPool)
			dealInfo.FeeToStockReserve = dealInfo.FeeToStockReserve.Add(stockToPool)
		} else {
			if takerRebateAmount, takerOrderReferenceAddr, err = pk.AllocateFeeToValidator(ctx, currOrder.Money(), moneyFee, currOrder.Sender); err != nil {
				panic(err)
			}
			if makerRebateAmount, makerOrderReferenceAddr, err = pk.AllocateFeeToValidator(ctx, currOrder.Stock(), stockFee, orderInBook.Sender); err != nil {
				panic(err)
			}
		}
	} else {
		takerFee = moneyFee
		makerFee = stockFee
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Stock(), stockTrans.Sub(stockFee))
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Money(), moneyTrans.Sub(moneyFee))
		if isPoolExists {
			stockToPool, takerRebateAmount, takerOrderReferenceAddr, err = pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Stock(), stockFee, currOrder.Sender)
			if err != nil {
				panic(err)
			}
			moneyToPool, makerRebateAmount, makerOrderReferenceAddr, err = pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Money(), moneyFee, orderInBook.Sender)
			if err != nil {
				panic(err)
			}
			dealInfo.FeeToMoneyReserve = dealInfo.FeeToMoneyReserve.Add(moneyToPool)
			dealInfo.FeeToStockReserve = dealInfo.FeeToStockReserve.Add(stockToPool)
		} else {
			if takerRebateAmount, takerOrderReferenceAddr, err = pk.AllocateFeeToValidator(ctx, currOrder.Stock(), stockFee, currOrder.Sender); err != nil {
				panic(err)
			}
			if makerRebateAmount, makerOrderReferenceAddr, err = pk.AllocateFeeToValidator(ctx, currOrder.Money(), moneyFee, orderInBook.Sender); err != nil {
				panic(err)
			}
		}
	}
	pk.sendDealOrderMsg(ctx, currOrder, orderInBook, dealStockAmount, moneyTrans.Int64(),
		takerFee.Int64(), makerFee.Int64(), poolInfo, takerRebateAmount, makerRebateAmount, takerOrderReferenceAddr, makerOrderReferenceAddr)
}

func (pk PairKeeper) sendDealOrderMsg(ctx sdk.Context, order, dealOrder *types.Order, dealStock, dealMoney int64,
	takerFee, makerFee int64, poolInfo *PoolInfo, takerRebateAmount, makerRebateAmount sdk.Int, takerRebateReferenceAddr, makerRebateReferenceAddr sdk.AccAddress) {
	if pk.msgProducer == nil || pk.msgProducer.IsSubscribed(types.ModuleName) {
		return
	}
	taker := types.FillOrderInfoMq{
		OrderID:     order.GetOrderID(),
		TradingPair: order.TradingPair,
		Height:      ctx.BlockHeight(),
		Price:       order.Price,
		Side:        getSide(order.IsBuy),

		LeftStock:              order.LeftStock,
		Freeze:                 order.Freeze,
		DealStock:              order.DealStock,
		DealMoney:              order.DealMoney,
		CurrMoney:              dealMoney,
		CurrStock:              dealStock,
		FillPrice:              dealOrder.Price,
		CurrUsedCommission:     takerFee,
		TakerRebateAmount:      takerRebateAmount.Int64(),
		MakerRebateAmount:      makerRebateAmount.Int64(),
		TakerRebateRefereeAddr: takerRebateReferenceAddr.String(),
		MakerRebateRefereeAddr: makerRebateReferenceAddr.String(),
	}
	maker := types.FillOrderInfoMq{
		OrderID:     dealOrder.GetOrderID(),
		TradingPair: dealOrder.TradingPair,
		Height:      ctx.BlockHeight(),
		Price:       dealOrder.Price,
		Side:        getSide(dealOrder.IsBuy),

		LeftStock:              dealOrder.LeftStock,
		Freeze:                 dealOrder.Freeze,
		DealStock:              dealOrder.DealStock,
		DealMoney:              dealOrder.DealMoney,
		CurrMoney:              dealMoney,
		CurrStock:              dealStock,
		FillPrice:              dealOrder.Price,
		CurrUsedCommission:     makerFee,
		TakerRebateAmount:      takerRebateAmount.Int64(),
		MakerRebateAmount:      makerRebateAmount.Int64(),
		TakerRebateRefereeAddr: takerRebateReferenceAddr.String(),
		MakerRebateRefereeAddr: makerRebateReferenceAddr.String(),
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
	rebateAmount, referenceAddr, fee, poolToUser := pk.dealWithPoolAndCollectFee(ctx, order, dealInfo, poolInfo)
	if dealInfo.AmountInToPool.IsPositive() {
		pk.sendDealInfoWithPool(ctx, dealInfo, order, fee.Int64(), poolToUser.Int64(), poolInfo, rebateAmount, referenceAddr)
	}
	poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Add(dealInfo.FeeToMoneyReserve)
	poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Add(dealInfo.FeeToStockReserve)
}

func (pk PairKeeper) dealWithPoolAndCollectFee(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) (rebateAmount sdk.Int, referenceAddr sdk.AccAddress, fee sdk.Int, poolToUser sdk.Int) {
	outAmount := sdk.ZeroInt()
	if dealInfo.AmountInToPool.IsPositive() {
		outAmount = GetAmountOutInPool(dealInfo.AmountInToPool, poolInfo, order.IsBuy)
	} else {
		return sdk.ZeroInt(), nil, sdk.ZeroInt(), sdk.ZeroInt()
	}
	// add fee calculate
	feeRate := pk.GetParams(ctx).DealWithPoolFeeRate
	fee = outAmount.Mul(sdk.NewInt(feeRate)).Add(sdk.NewInt(9999)).Quo(sdk.NewInt(types.DefaultFeePrecision))
	if order.IsBuy {
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Sub(outAmount)
	} else {
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Sub(outAmount)
	}
	var (
		err         sdk.Error
		stockToPool sdk.Int
		moneyToPool sdk.Int
	)
	if order.IsBuy {
		if err := pk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, order.Sender, newCoins(order.Stock(), outAmount)); err != nil {
			panic(err)
		}
		stockToPool, rebateAmount, referenceAddr, err = pk.AllocateFeeToValidatorAndPool(ctx, order.Stock(), fee, order.Sender)
		if err != nil {
			panic(err)
		}
		dealInfo.FeeToStockReserve = dealInfo.FeeToStockReserve.Add(stockToPool)
		if err := pk.UnFreezeCoins(ctx, order.Sender, newCoins(order.Money(), dealInfo.AmountInToPool)); err != nil {
			panic(err)
		}
		if err := pk.SendCoinsFromAccountToModule(ctx, order.Sender, types.PoolModuleAcc, newCoins(order.Money(), dealInfo.AmountInToPool)); err != nil {
			panic(err)
		}
	} else {
		if err := pk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, order.Sender, newCoins(order.Money(), outAmount)); err != nil {
			panic(err)
		}
		moneyToPool, rebateAmount, referenceAddr, err = pk.AllocateFeeToValidatorAndPool(ctx, order.Money(), fee, order.Sender)
		if err != nil {
			panic(err)
		}
		dealInfo.FeeToMoneyReserve = dealInfo.FeeToMoneyReserve.Add(moneyToPool)
		if err := pk.UnFreezeCoins(ctx, order.Sender, newCoins(order.Stock(), dealInfo.AmountInToPool)); err != nil {
			panic(err)
		}
		if err := pk.SendCoinsFromAccountToModule(ctx, order.Sender, types.PoolModuleAcc, newCoins(order.Stock(), dealInfo.AmountInToPool)); err != nil {
			panic(err)
		}
	}
	return rebateAmount, referenceAddr, fee, outAmount
}

func (pk PairKeeper) sendDealInfoWithPool(ctx sdk.Context, dealInfo *types.DealInfo,
	order *types.Order, commission, poolAmount int64, poolInfo *PoolInfo, rebateAmount sdk.Int, referenceAddr sdk.AccAddress) {
	if pk.msgProducer == nil || pk.msgProducer.IsSubscribed(types.ModuleName) {
		return
	}
	currStock := poolAmount
	currMoney := dealInfo.AmountInToPool.Int64()
	if !order.IsBuy {
		currStock = dealInfo.AmountInToPool.Int64()
		currMoney = poolAmount
	}
	dealOrderInfo := types.FillOrderInfoWithPoolMq{
		OrderID:            order.GetOrderID(),
		TradingPair:        order.TradingPair,
		Height:             ctx.BlockHeight(),
		Side:               getSide(order.IsBuy),
		Price:              order.Price,
		Freeze:             order.Freeze,
		LeftStock:          order.LeftStock,
		FillPrice:          sdk.NewDec(currMoney).Quo(sdk.NewDec(currStock)),
		DealAmountWithPool: poolAmount,
		CurrUsedCommission: commission,
		RebateAmount:       rebateAmount.Int64(),
		RebateRefereeAddr:  referenceAddr.String(),
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
	return pk.IOrderBookKeeper.GetOrder(ctx, &market.QueryOrderParam{
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
		info.MoneyOrderBookReserve = info.MoneyOrderBookReserve.Sub(sdk.NewInt(delOrder.Freeze))
		if err := pk.UnFreezeCoins(ctx, delOrder.Sender, newCoins(delOrder.Money(), sdk.NewInt(delOrder.Freeze))); err != nil {
			return err
		}
	} else {
		info.StockOrderBookReserve = info.StockOrderBookReserve.Sub(sdk.NewInt(delOrder.Freeze))
		if err := pk.UnFreezeCoins(ctx, delOrder.Sender, newCoins(delOrder.Stock(), sdk.NewInt(delOrder.Freeze))); err != nil {
			return err
		}
	}
	pk.SetPoolInfo(ctx, delOrder.TradingPair, info)
	return nil
}

func (pk *PairKeeper) GetAllOrders(ctx sdk.Context, market string) []*types.Order {
	return pk.GetAllOrdersInMarket(ctx, market)
}

func (pk *PairKeeper) GetOrdersFromUser(ctx sdk.Context, user string) []string {
	return pk.IOrderBookKeeper.GetOrdersFromUser(ctx, user)
}
