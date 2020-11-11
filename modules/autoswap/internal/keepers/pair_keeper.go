package keepers

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
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
	AddMarketOrder(ctx sdk.Context, order *types.Order) (sdk.Int, sdk.Error)
	AddLimitOrder(ctx sdk.Context, order *types.Order) sdk.Error
	GetFirstOrderID(ctx sdk.Context, symbol string, isBuy bool) int64
	DeleteOrder(ctx sdk.Context, order *types.MsgDeleteOrder) sdk.Error
	HasOrder(ctx sdk.Context, symbol string, isBuy bool, orderID int64) bool
	GetOrder(ctx sdk.Context, symbol string, isBuy bool, orderID int64) *types.Order
	SetParams(ctx sdk.Context, params types.Params)
	GetParams(ctx sdk.Context) types.Params
	GetTakerFee(ctx sdk.Context) sdk.Dec
	GetMakerFee(ctx sdk.Context) sdk.Dec
	GetDealWithPoolFee(ctx sdk.Context) sdk.Dec
	GetFeeToValidator(ctx sdk.Context) sdk.Dec

	GetPairList() map[Pair]struct{}
	ClearPairList()
}

type FeeFunc func(sdk.Context) sdk.Dec

type PairKeeper struct {
	IPoolKeeper
	types.SupplyKeeper
	types.ExpectedBankKeeper
	codec    *codec.Codec
	storeKey sdk.StoreKey
	subspace params.Subspace
	// record deal pairs in one block.
	DealPairs map[Pair]struct{}
}

func NewPairKeeper(poolKeeper IPoolKeeper, supplyK types.SupplyKeeper, bnk types.ExpectedBankKeeper,
	codec *codec.Codec, storeKey sdk.StoreKey, paramSubspace params.Subspace) *PairKeeper {
	return &PairKeeper{
		codec:              codec,
		storeKey:           storeKey,
		IPoolKeeper:        poolKeeper,
		SupplyKeeper:       supplyK,
		ExpectedBankKeeper: bnk,
		subspace:           paramSubspace.WithKeyTable(types.ParamKeyTable()),

		DealPairs: make(map[Pair]struct{}),
	}
}

func (pk *PairKeeper) GetPairList() map[Pair]struct{} {
	return pk.DealPairs
}
func (pk *PairKeeper) ClearPairList() {
	pk.DealPairs = make(map[Pair]struct{})
}

func (keeper *PairKeeper) SetParams(ctx sdk.Context, params types.Params) {
	keeper.subspace.SetParamSet(ctx, &params)
}

func (keeper *PairKeeper) GetParams(ctx sdk.Context) (param types.Params) {
	keeper.subspace.GetParamSet(ctx, &param)
	return
}

func (keeper *PairKeeper) GetTakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(keeper.GetParams(ctx).TakerFeeRateRate).QuoInt64(types.DefaultFeePrecision)
}

func (keeper *PairKeeper) GetMakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(keeper.GetParams(ctx).MakerFeeRateRate).QuoInt64(types.DefaultFeePrecision)
}

func (keeper *PairKeeper) GetDealWithPoolFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(keeper.GetParams(ctx).DealWithPoolFeeRate).QuoInt64(types.DefaultFeePrecision)
}

func (keeper *PairKeeper) GetFeeToValidator(ctx sdk.Context) sdk.Dec {
	param := keeper.GetParams(ctx)
	return sdk.NewDec(param.FeeToValidator).QuoInt64(param.FeeToValidator + param.FeeToPool)
}

func (keeper *PairKeeper) AllocateFeeToValidatorAndPool(ctx sdk.Context, denom string, totalAmount sdk.Int, sender sdk.AccAddress) sdk.Error {
	feeToVal := keeper.GetFeeToValidator(ctx).MulInt(totalAmount).TruncateInt()
	feeToPool := totalAmount.Sub(feeToVal)
	err := keeper.SendCoinsFromAccountToModule(ctx, sender, auth.FeeCollectorName, sdk.NewCoins(sdk.NewCoin(denom, feeToVal)))
	if err != nil {
		return err
	}
	err = keeper.SendCoinsFromAccountToModule(ctx, sender, types.PoolModuleAcc, sdk.NewCoins(sdk.NewCoin(denom, feeToPool)))
	if err != nil {
		return err
	}
	return nil
}

func (keeper *PairKeeper) AllocateFeeToValidator(ctx sdk.Context, fee sdk.Coins, sender sdk.AccAddress) sdk.Error {
	err := keeper.SendCoinsFromAccountToModule(ctx, sender, auth.FeeCollectorName, fee)
	if err != nil {
		return err
	}
	return nil
}

func (pk PairKeeper) AddLimitOrder(ctx sdk.Context, order *types.Order) (err sdk.Error) {
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

	var poolInfo *PoolInfo
	if poolInfo = pk.GetPoolInfo(ctx, order.MarketSymbol); poolInfo == nil {
		return types.ErrInvalidMarket(order.MarketSymbol)
	}
	if order.OrderID = pk.getUnusedOrderID(ctx, order); order.OrderID <= 0 {
		return types.ErrInvalidOrderID(order.OrderID)
	}

	//1. will calculate and check order amount
	// freeze order balance in sender account
	actualAmount, err := pk.freezeOrderCoin(ctx, order)
	if err != nil {
		return err
	}

	//2. calculate insert order position, try insert order if the order can't deal.
	if order.HasPrevKey() {
		if pk.insertOrderFromGivePos(ctx, order) {
			//3. when insert order later, update pool info
			if order.IsBuy {
				poolInfo.MoneyOrderBookReserve = poolInfo.MoneyOrderBookReserve.Add(actualAmount)
			} else {
				poolInfo.StockOrderBookReserve = poolInfo.StockOrderBookReserve.Add(actualAmount)
			}
			pk.SetPoolInfo(ctx, order.MarketSymbol, poolInfo)
			return nil
		}
	}
	//4. deal the order and insert remained order to orderBook.
	amountToTaker, err := pk.dealOrderAndAddRemainedOrder(ctx, order, poolInfo)
	if err != nil {
		return err
	}
	if amountToTaker.IsPositive() {
		pk.DealPairs[Pair{
			Symbol: order.MarketSymbol,
		}] = struct{}{}
	}
	return nil
}

func (pk PairKeeper) getUnusedOrderID(ctx sdk.Context, order *types.Order) int64 {
	var id = order.OrderID
	if order.OrderID == 0 {
		id = int64(binary.BigEndian.Uint64(ctx.BlockHeader().AppHash[:8]) ^
			binary.BigEndian.Uint64(order.Sender[:8]))
	}

	for i := 0; i < 100 && id > 0; i++ {
		if pk.HasOrder(ctx, order.MarketSymbol, order.IsBuy, id) {
			id++
			continue
		}
		return id
	}
	return -1
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

// insertOrderFromGivePos will insert order to given position
// If insert success, return ture. Otherwise, return false
func (pk PairKeeper) insertOrderFromGivePos(ctx sdk.Context, order *types.Order) bool {
	var prevOrder *types.Order
	if prevOrder = pk.getPrevOrder3Times(ctx, order); prevOrder == nil {
		return false
	}
	return pk.insertOrder(ctx, order, prevOrder)
}

func (pk PairKeeper) getPrevOrder3Times(ctx sdk.Context, order *types.Order) (prevOrder *types.Order) {
	for _, v := range order.PrevKey {
		if prevOrder = pk.GetOrder(ctx, order.MarketSymbol, order.IsBuy, v); prevOrder != nil {
			return prevOrder
		}
	}
	return nil
}

func (pk PairKeeper) insertOrder(ctx sdk.Context, order, prevOrder *types.Order) bool {
	prevOrderId := prevOrder.OrderID
	var nextOrder *types.Order
	for prevOrderId > 0 {
		canFollow := (order.IsBuy && (order.Price.LTE(prevOrder.Price))) ||
			(!order.IsBuy && (order.Price.GTE(prevOrder.Price)))
		if !canFollow {
			break
		}
		if prevOrder.NextOrderID > 0 {
			nextOrder = pk.GetOrder(ctx, prevOrder.MarketSymbol, prevOrder.IsBuy, prevOrder.NextOrderID)
			canPrecede := (order.IsBuy && (order.Price.GT(nextOrder.Price))) ||
				(!order.IsBuy && (order.Price.LT(nextOrder.Price)))
			canFollow = canFollow && canPrecede
		}
		if canFollow {
			order.NextOrderID = prevOrder.NextOrderID
			pk.SetOrder(ctx, order)
			prevOrder.NextOrderID = order.OrderID
			pk.SetOrder(ctx, prevOrder)
			return true
		}
		prevOrderId = prevOrder.NextOrderID
		prevOrder = nextOrder
	}
	return false
}

// dealOrderAndAddRemainedOrder Deal the order and
// insert the remainder order into the order book
func (pk PairKeeper) dealOrderAndAddRemainedOrder(ctx sdk.Context, order *types.Order, poolInfo *PoolInfo) (sdk.Int, sdk.Error) {
	firstOrderID := pk.GetFirstOrderID(ctx, order.MarketSymbol, !order.IsBuy)
	currOrderID := firstOrderID
	dealInfo := &types.DealInfo{
		RemainAmount:    order.ActualAmount(),
		AmountInToPool:  sdk.ZeroInt(),
		DealMoneyInBook: sdk.ZeroInt(),
		DealStockInBook: sdk.ZeroInt(),
	}
	if !order.IsLimitOrder {
		dealInfo.RemainAmount = order.Amount
	}
	for currOrderID > 0 {
		orderInBook := pk.GetOrder(ctx, order.MarketSymbol, !order.IsBuy, currOrderID)
		canDealInBook := (order.IsBuy && order.Price.GTE(orderInBook.Price)) ||
			(!order.IsBuy && order.Price.LTE(orderInBook.Price))
		// can't deal with order book
		if !canDealInBook {
			break
		}
		// full deal in pool
		if allDeal := pk.tryDealInPool(dealInfo, orderInBook.Price, order.IsBuy, poolInfo); allDeal {
			break
		}
		// deal in order book
		isPoolExists := true
		if poolInfo == nil || poolInfo.StockAmmReserve.IsZero() {
			isPoolExists = false
		}
		pk.dealInOrderBook(ctx, order, orderInBook, dealInfo, isPoolExists)

		// the order in order book didn't fully deal, then the new order did fully deal.
		// update remained order info to order book.
		if orderInBook.Amount.IsPositive() {
			pk.SetOrder(ctx, orderInBook)
			break
		}
		// the order in order book have fully deal, so delete the order info.
		// update the curr order id that will deal next round.
		pk.deleteOrder(ctx, orderInBook)
		currOrderID = orderInBook.NextOrderID
	}

	if order.IsLimitOrder {
		pk.tryDealInPool(dealInfo, order.Price, order.IsBuy, poolInfo)
		pk.insertOrderToBook(ctx, order, dealInfo, poolInfo)
	} else {
		dealInfo.AmountInToPool = dealInfo.AmountInToPool.Add(order.ActualAmount())
		dealInfo.RemainAmount = sdk.ZeroInt()
	}
	amountToTaker := pk.dealWithPoolAndCollectFee(ctx, order, dealInfo, poolInfo)
	if order.IsBuy {
		poolInfo.StockOrderBookReserve = poolInfo.StockOrderBookReserve.Sub(dealInfo.DealStockInBook)
	} else {
		poolInfo.MoneyOrderBookReserve = poolInfo.MoneyOrderBookReserve.Sub(dealInfo.DealMoneyInBook)
	}
	if firstOrderID != currOrderID {
		pk.SetFirstOrderID(ctx, order.MarketSymbol, !order.IsBuy, currOrderID)
	}
	pk.SetPoolInfo(ctx, order.MarketSymbol, poolInfo)
	return amountToTaker, nil
}

func (pk PairKeeper) tryDealInPool(dealInfo *types.DealInfo, dealPrice sdk.Dec, isBuy bool, info *PoolInfo) bool {
	currTokenCanTradeWithPool := pk.intoPoolAmountTillPrice(dealPrice, isBuy, info)
	// will check deal token amount later.
	if currTokenCanTradeWithPool.GT(types.MaxAmount) {
		panic("deal amount with pool is too large")
	}
	if currTokenCanTradeWithPool.GT(dealInfo.AmountInToPool) {
		diffTokenTradeWithPool := currTokenCanTradeWithPool.Sub(dealInfo.AmountInToPool)
		allDeal := diffTokenTradeWithPool.GT(dealInfo.RemainAmount)
		if allDeal {
			diffTokenTradeWithPool = dealInfo.RemainAmount
		}
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(diffTokenTradeWithPool)
		dealInfo.AmountInToPool = dealInfo.AmountInToPool.Add(diffTokenTradeWithPool)
		return allDeal
	}
	return false
}

func (pk PairKeeper) intoPoolAmountTillPrice(dealPrice sdk.Dec, isBuy bool, info *PoolInfo) sdk.Int {
	if isBuy {
		root := dealPrice.Mul(sdk.NewDecFromInt(info.StockAmmReserve)).Mul(sdk.NewDecFromInt(info.MoneyAmmReserve))
		return sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.Int)).Sub(sdk.NewDecFromInt(info.MoneyAmmReserve)).TruncateInt()
	}
	root := sdk.NewDecFromInt(info.StockAmmReserve).Mul(sdk.NewDecFromInt(info.MoneyAmmReserve)).Quo(dealPrice)
	return sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.Int)).Sub(sdk.NewDecFromInt(info.StockAmmReserve)).TruncateInt()
}

func (pk PairKeeper) dealInOrderBook(ctx sdk.Context, currOrder, orderInBook *types.Order, dealInfo *types.DealInfo, isPoolExists bool) {
	dealInfo.HasDealInOrderBook = true
	stockAmount := sdk.Int{}
	// will calculate stock amount
	if currOrder.IsBuy {
		//the stock amount might be smaller than actual
		stockAmount = sdk.NewDecFromInt(dealInfo.RemainAmount).Quo(currOrder.Price).TruncateInt()
	} else {
		stockAmount = dealInfo.RemainAmount
	}
	if orderInBook.Amount.LT(stockAmount) {
		stockAmount = orderInBook.Amount
	}
	orderInBook.Amount = orderInBook.Amount.Sub(stockAmount)
	currOrder.Amount = currOrder.Amount.Sub(stockAmount)

	var (
		stockFee   sdk.Int
		moneyFee   sdk.Int
		stockTrans = stockAmount
		moneyTrans = sdk.NewDecFromInt(stockAmount).Mul(orderInBook.Price).TruncateInt()
	)
	if currOrder.IsBuy {
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(moneyTrans)
		moneyFee = pk.GetMakerFee(ctx).MulInt(moneyTrans).TruncateInt()
		stockFee = pk.GetTakerFee(ctx).MulInt(stockTrans).TruncateInt()
	} else {
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(stockTrans)
		stockFee = pk.GetMakerFee(ctx).MulInt(stockTrans).TruncateInt()
		moneyFee = pk.GetTakerFee(ctx).MulInt(moneyTrans).TruncateInt()
	}

	dealInfo.DealMoneyInBook = dealInfo.DealMoneyInBook.Add(moneyTrans)
	dealInfo.DealStockInBook = dealInfo.DealStockInBook.Add(stockTrans)
	// transfer tokens in orders
	if currOrder.IsBuy {
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Money(), moneyTrans.Sub(moneyFee))
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Stock(), stockTrans.Sub(stockFee))
		if isPoolExists {
			if err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Money(), moneyFee, currOrder.Sender); err != nil {
				panic(err)
			}
			if err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Stock(), stockFee, orderInBook.Sender); err != nil {
				panic(err)
			}
		} else {
			if err := pk.AllocateFeeToValidator(ctx, sdk.NewCoins(sdk.NewCoin(currOrder.Money(), moneyFee),
				sdk.NewCoin(currOrder.Stock(), stockFee)), currOrder.Sender); err != nil {
				panic(err)
			}
		}

	} else {
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Stock(), stockTrans.Sub(stockFee))
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Money(), moneyTrans.Sub(moneyFee))
		if err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Stock(), stockFee, currOrder.Sender); err != nil {
			panic(err)
		}
		if err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Money(), moneyFee, orderInBook.Sender); err != nil {
			panic(err)
		}
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

func (pk PairKeeper) insertOrderToBook(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) {
	var (
		moneyAmount sdk.Int
		stockAmount sdk.Int
	)
	if order.IsBuy {
		moneyAmount = dealInfo.RemainAmount
		stockAmount = sdk.NewDecFromInt(dealInfo.RemainAmount).Quo(order.Price).TruncateInt()
	} else {
		stockAmount = dealInfo.RemainAmount
	}
	if stockAmount.IsPositive() {
		order.Amount = stockAmount
		if dealInfo.HasDealInOrderBook {
			pk.insertOrderAtHead(ctx, order)
		} else {
			pk.insertOrderFromHead(ctx, order)
		}
	}
	if order.IsBuy {
		poolInfo.MoneyOrderBookReserve = poolInfo.MoneyOrderBookReserve.Add(moneyAmount)
	} else {
		poolInfo.StockOrderBookReserve = poolInfo.StockOrderBookReserve.Add(stockAmount)
	}
}

func (pk PairKeeper) insertOrderAtHead(ctx sdk.Context, order *types.Order) {
	order.NextOrderID = pk.GetFirstOrderID(ctx, order.MarketSymbol, order.IsBuy)
	pk.SetOrder(ctx, order)
	pk.SetFirstOrderID(ctx, order.MarketSymbol, order.IsBuy, order.OrderID)
}

func (pk PairKeeper) insertOrderFromHead(ctx sdk.Context, order *types.Order) {
	firstOrderID := pk.GetFirstOrderID(ctx, order.MarketSymbol, order.IsBuy)
	var (
		firstOrder *types.Order
		canBeFirst = firstOrderID <= 0
	)
	if !canBeFirst {
		firstOrder = pk.GetOrder(ctx, order.MarketSymbol, order.IsBuy, firstOrderID)
		canBeFirst = (order.IsBuy && order.Price.GT(firstOrder.Price)) ||
			(!order.IsBuy && order.Price.LT(firstOrder.Price))
	}
	if canBeFirst {
		order.NextOrderID = firstOrderID
		pk.SetOrder(ctx, order)
		pk.SetFirstOrderID(ctx, order.MarketSymbol, order.IsBuy, order.OrderID)
		return
	}
	pk.insertOrder(ctx, order, firstOrder)
}

func (pk PairKeeper) dealWithPoolAndCollectFee(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) sdk.Int {
	outPoolTokenReserve, inPoolTokenReserve, otherToTaker := poolInfo.MoneyAmmReserve, poolInfo.StockAmmReserve, dealInfo.DealMoneyInBook
	if order.IsBuy {
		outPoolTokenReserve, inPoolTokenReserve, otherToTaker = poolInfo.StockAmmReserve, poolInfo.MoneyAmmReserve, dealInfo.DealStockInBook
	}
	outAmount := sdk.ZeroInt()
	if !dealInfo.AmountInToPool.IsZero() {
		outAmount = outPoolTokenReserve.Mul(dealInfo.AmountInToPool).Quo(inPoolTokenReserve.Add(dealInfo.AmountInToPool))
	}
	if dealInfo.AmountInToPool.IsPositive() {
		// todo emit deal with pool log
	}

	// add fee calculate
	fee := pk.GetDealWithPoolFee(ctx).MulInt(outAmount).TruncateInt()
	amountToTaker := outAmount.Add(otherToTaker).Sub(fee)
	outAmount = outAmount.Sub(fee)
	if order.IsBuy {
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Sub(outAmount)
	} else {
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Sub(outAmount)
	}
	// transfer token from pool to order sender
	if order.IsBuy {
		if err := pk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, order.Sender, newCoins(order.Stock(), outAmount)); err != nil {
			panic(err)
		}
		if err := pk.AllocateFeeToValidatorAndPool(ctx, order.Money(), dealInfo.AmountInToPool, order.Sender); err != nil {
			panic(err)
		}
	} else {
		fmt.Println(newCoins(order.Money(), outAmount).String())
		if err := pk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, order.Sender, newCoins(order.Money(), outAmount)); err != nil {
			panic(err)
		}
		if err := pk.AllocateFeeToValidatorAndPool(ctx, order.Stock(), dealInfo.AmountInToPool, order.Sender); err != nil {
			panic(err)
		}
	}
	return amountToTaker
}

func (pk PairKeeper) AddMarketOrder(ctx sdk.Context, order *types.Order) (sdk.Int, sdk.Error) {
	order.Price = sdk.ZeroDec()
	if order.IsBuy {
		order.Price = sdk.NewDec(math.MaxInt64)
	}
	order.IsLimitOrder = false
	poolInfo := pk.GetPoolInfo(ctx, order.MarketSymbol)
	amountToTaker, err := pk.dealOrderAndAddRemainedOrder(ctx, order, poolInfo)
	if err != nil {
		return sdk.Int{}, err
	}
	if amountToTaker.LT(order.MinOutputAmount) {
		return sdk.Int{}, types.ErrAmountOutIsSmallerThanExpected(order.MinOutputAmount, amountToTaker)
	}
	if amountToTaker.IsPositive() {
		pk.DealPairs[Pair{
			Symbol: order.MarketSymbol,
		}] = struct{}{}
	}
	return amountToTaker, nil
}

func (pk PairKeeper) HasOrder(ctx sdk.Context, symbol string, isBuy bool, orderID int64) bool {
	return pk.GetOrder(ctx, symbol, isBuy, orderID) != nil
}

func (pk PairKeeper) GetOrder(ctx sdk.Context, symbol string, isBuy bool, orderID int64) *types.Order {
	var recordOrder *types.Order
	store := ctx.KVStore(pk.storeKey)
	key := getOrderKey(&types.Order{
		OrderBasic: types.OrderBasic{
			MarketSymbol: symbol,
			IsBuy:        isBuy,
		},
		OrderID: orderID,
	})
	val := store.Get(key)
	if len(val) == 0 {
		return nil
	}
	pk.codec.MustUnmarshalBinaryBare(val, &recordOrder)
	return recordOrder
}

func (pk PairKeeper) SetOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(pk.storeKey)
	key := getOrderKey(order)
	val := pk.codec.MustMarshalBinaryBare(order)
	store.Set(key, val)
}

func (pk PairKeeper) DeleteOrder(ctx sdk.Context, delOrder *types.MsgDeleteOrder) sdk.Error {
	order := pk.GetOrder(ctx, delOrder.MarketSymbol, delOrder.IsBuy, delOrder.OrderID)
	if order == nil || order.Sender.Equals(delOrder.Sender) {
		var sender sdk.AccAddress
		if order == nil {
			sender = sdk.AccAddress{}
		} else {
			sender = order.Sender
		}
		return types.ErrInvalidOrderNews(fmt.Sprintf("market: %s, isBuy: %v, orderID: %d, orderSender: %s",
			delOrder.MarketSymbol, delOrder.IsBuy, delOrder.OrderID, sender.String()))
	}
	if !order.HasPrevKey() {
		firstOrderID := pk.GetFirstOrderID(ctx, delOrder.MarketSymbol, delOrder.IsBuy)
		if firstOrderID != order.OrderID {
			return types.ErrNotFoundOrder(fmt.Sprintf("the delete order is not the first in the orderBook: currentID: %d, "+
				"firstOrderID: %d, shoule fill the prev_key in MsgDeleteOrder", delOrder.OrderID, firstOrderID))
		}
		pk.SetFirstOrderID(ctx, order.MarketSymbol, order.IsBuy, order.NextOrderID)
	} else {
		orderInBook := pk.getPrevOrder3Times(ctx, order)
		if orderInBook == nil {
			return types.ErrInvalidPrevKey(order.PrevKey)
		}
		currentOrderID := orderInBook.OrderID
		for currentOrderID > 0 {
			if orderInBook.NextOrderID == order.OrderID {
				orderInBook.NextOrderID = order.NextOrderID
				pk.SetOrder(ctx, orderInBook)
				break
			}
			orderInBook = pk.GetOrder(ctx, order.MarketSymbol, order.IsBuy, orderInBook.NextOrderID)
			if orderInBook == nil {
				return types.ErrNotFoundOrder(fmt.Sprintf("the delete order not found by prev_key in MsgDeleteOrder, receive"+
					" data is: %v", delOrder.PrevKey))
			}
			currentOrderID = orderInBook.OrderID
		}
	}
	return pk.updateOrderBookReserveByOrderDel(ctx, order)
}

func (pk PairKeeper) updateOrderBookReserveByOrderDel(ctx sdk.Context, delOrder *types.Order) sdk.Error {
	info := pk.GetPoolInfo(ctx, delOrder.MarketSymbol)
	amount := delOrder.Amount
	if delOrder.IsBuy {
		amount = delOrder.ActualAmount()
		info.MoneyOrderBookReserve = info.MoneyOrderBookReserve.Sub(amount)
		if err := pk.UnFreezeCoins(ctx, delOrder.Sender, newCoins(delOrder.Money(), delOrder.Amount)); err != nil {
			return err
		}
	} else {
		info.StockOrderBookReserve = info.StockOrderBookReserve.Sub(delOrder.Amount)
		if err := pk.UnFreezeCoins(ctx, delOrder.Sender, newCoins(delOrder.Stock(), delOrder.Amount)); err != nil {
			return err
		}
	}
	pk.SetPoolInfo(ctx, delOrder.MarketSymbol, info)
	return nil
}

func (pk PairKeeper) deleteOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(pk.storeKey)
	key := getOrderKey(order)
	store.Delete(key)
}

func (pk PairKeeper) GetFirstOrderID(ctx sdk.Context, symbol string, isBuy bool) int64 {
	store := ctx.KVStore(pk.storeKey)
	key := getBestOrderPriceKey(symbol, isBuy)
	val := store.Get(key)
	if len(val) == 0 {
		return -1
	}
	return int64(binary.BigEndian.Uint64(val))
}

func (pk PairKeeper) SetFirstOrderID(ctx sdk.Context, symbol string, isBuy bool, orderID int64) {
	store := ctx.KVStore(pk.storeKey)
	key := getBestOrderPriceKey(symbol, isBuy)
	val := make([]byte, 8)
	binary.BigEndian.PutUint64(val, uint64(orderID))
	store.Set(key, val)
}
