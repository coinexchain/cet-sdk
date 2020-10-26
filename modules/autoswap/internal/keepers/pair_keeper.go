package keepers

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IPairKeeper interface {
	IPoolKeeper
	AddMarketOrder(ctx sdk.Context, order *types.Order) bool
	AddLimitOrder(ctx sdk.Context, order *types.Order) bool
	HasOrder(ctx sdk.Context, isBuy bool, orderID uint64) bool
	GetOrder(ctx sdk.Context, isBuy bool, orderID uint64) *types.Order
}

type PairKeeper struct {
	IPoolKeeper
	types.ExpectedBankKeeper
	types.ExpectedAccountKeeper
	codec    *codec.Codec
	storeKey sdk.StoreKey
}

func NewPairKeeper(poolKeeper IPoolKeeper, bnk types.ExpectedBankKeeper,
	acck types.ExpectedAccountKeeper, codec *codec.Codec, storeKey sdk.StoreKey) *PairKeeper {
	return &PairKeeper{
		codec:                 codec,
		storeKey:              storeKey,
		IPoolKeeper:           poolKeeper,
		ExpectedBankKeeper:    bnk,
		ExpectedAccountKeeper: acck,
	}
}

func (pk PairKeeper) AddLimitOrder(ctx sdk.Context, order *types.Order) (err error) {
	defer func() {
		r := recover()
		switch r.(type) {
		case error:
			err = r.(error)
		case string:
			err = sdk.NewError(types.RouterKey, 3, r.(string))
		}
	}()

	var poolInfo *PoolInfo
	if poolInfo = pk.GetPoolInfo(ctx, order.MarketSymbol, order.IsOpenSwap); poolInfo == nil {
		return fmt.Errorf("can't find pool info")
	}
	if order.OrderID = pk.getUnusedOrderID(ctx, order); order.OrderID <= 0 {
		return fmt.Errorf("can't find available order id")
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
				poolInfo.moneyOrderBookReserve.Add(actualAmount)
			} else {
				poolInfo.stockOrderBookReserve.Add(actualAmount)
			}
			pk.SetPoolInfo(ctx, order.MarketSymbol, order.IsOpenSwap, poolInfo)
			return nil
		}
	}
	//4. deal the order and insert remained order to orderBook.
	if _, err := pk.dealOrderAndAddRemainedOrder(ctx, order, poolInfo); err != nil {
		return err
	}
	return nil
}

func (pk PairKeeper) getUnusedOrderID(ctx sdk.Context, order *types.Order) int64 {
	var id int64 = 0
	if order.OrderID == 0 {
		id = int64(binary.BigEndian.Uint64(ctx.BlockHeader().AppHash[:8]) ^
			binary.BigEndian.Uint64(order.Sender[:8]))
	}

	for i := 0; i < 100 && id > 0; i++ {
		if pk.HasOrder(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy, id) {
			id++
			continue
		}
		return id
	}
	return -1
}

func (pk PairKeeper) freezeOrderCoin(ctx sdk.Context, order *types.Order) (sdk.Int, error) {
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
		if prevOrder = pk.GetOrder(ctx, order.MarketSymbol,
			order.IsOpenSwap, order.IsBuy, v); prevOrder != nil {
			return prevOrder
		}
	}
	return nil
}

func (pk PairKeeper) insertOrder(ctx sdk.Context, order, prevOrder *types.Order) bool {
	prevOrderId := prevOrder.OrderID
	var nextOrder *types.Order
	for prevOrderId > 0 {
		canFollow := (order.IsBuy && (order.Price <= prevOrder.Price)) ||
			(!order.IsBuy && (order.Price >= prevOrder.Price))
		if !canFollow {
			break
		}
		if prevOrder.NextOrderID > 0 {
			nextOrder = pk.GetOrder(ctx, prevOrder.MarketSymbol, prevOrder.IsOpenSwap, prevOrder.IsBuy, prevOrder.NextOrderID)
			canPrecede := (order.IsBuy && (order.Price > nextOrder.Price)) ||
				(!order.IsBuy && (order.Price < nextOrder.Price))
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
func (pk PairKeeper) dealOrderAndAddRemainedOrder(ctx sdk.Context, order *types.Order, poolInfo *PoolInfo) (sdk.Int, error) {
	firstOrderID := pk.GetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy)
	currOrderID := firstOrderID
	dealInfo := &types.DealInfo{RemainAmount: order.ActualAmount()}

	for currOrderID > 0 {
		orderInBook := pk.GetOrder(ctx, order.MarketSymbol, order.IsOpenSwap, !order.IsBuy, currOrderID)
		canDealInBook := (order.IsBuy && order.Price >= orderInBook.Price) ||
			(!order.IsBuy && order.Price <= orderInBook.Price)
		// can't deal with order book
		if !canDealInBook {
			break
		}
		// full deal in pool
		if allDeal := pk.tryDealInPool(dealInfo, orderInBook.ActualPrice(), order.IsBuy); allDeal {
			break
		}
		// deal in order book
		pk.dealInOrderBook(ctx, order, orderInBook, dealInfo)

		// the order in order book didn't fully deal, then the new order did fully deal.
		// update remained order info to order book.
		if orderInBook.Amount > 0 {
			pk.SetOrder(ctx, orderInBook)
			break
		}
		// the order in order book have fully deal, so delete the order info.
		// update the curr order id that will deal next round.
		pk.DeleteOrder(ctx, orderInBook)
		currOrderID = orderInBook.NextOrderID
	}

	if order.IsLimitOrder {
		pk.tryDealInPool(dealInfo, order.ActualPrice(), order.IsBuy)
		pk.insertOrderToBook(ctx, order, dealInfo, poolInfo)
	} else {
		dealInfo.AmountInToPool.Add(order.ActualAmount())
		dealInfo.RemainAmount = sdk.ZeroInt()
	}
	amountToTaker := pk.dealWithPoolAndCollectFee(ctx, order, dealInfo, poolInfo)
	if order.IsBuy {
		poolInfo.stockOrderBookReserve.Sub(dealInfo.DealStockInBook)
	} else {
		poolInfo.moneyOrderBookReserve.Sub(dealInfo.DealMoneyInBook)
	}
	if firstOrderID != currOrderID {
		pk.SetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, !order.IsBuy, currOrderID)
	}
	pk.SetPoolInfo(ctx, order.MarketSymbol, order.IsOpenSwap, poolInfo)
	return amountToTaker, nil
}

func (pk PairKeeper) tryDealInPool(dealInfo *types.DealInfo, dealPrice sdk.Dec, isBuy bool) bool {
	currTokenCanTradeWithPool := pk.intoPoolAmountTillPrice(dealPrice, isBuy)
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
		dealInfo.AmountInToPool.Add(diffTokenTradeWithPool)
		dealInfo.RemainAmount.Sub(diffTokenTradeWithPool)
		return allDeal
	}
	return false
}

func (pk PairKeeper) intoPoolAmountTillPrice(dealPrice sdk.Dec, isBuy bool) sdk.Int {
	return sdk.Int{}
}

func (pk PairKeeper) dealInOrderBook(ctx sdk.Context, currOrder, orderInBook *types.Order, dealInfo *types.DealInfo) {
	dealInfo.HasDealInOrderBook = true
	stockAmount := sdk.Int{}
	// will calculate stock amount
	if currOrder.IsBuy {
		//the stock amount might be smaller than actual
		stockAmount = sdk.NewDecFromInt(dealInfo.RemainAmount).Quo(currOrder.ActualPrice()).TruncateInt()
	} else {
		stockAmount = dealInfo.RemainAmount
	}
	if orderInBook.ActualAmount().LTE(stockAmount) {
		stockAmount = orderInBook.ActualAmount()
		orderInBook.Amount = 0
	} else {
		if orderInBook.IsBuy {
			orderInBook.Amount -= stockAmount.ToDec().Mul(orderInBook.ActualPrice()).TruncateInt().Int64()
		} else {
			orderInBook.Amount -= stockAmount.Int64()
		}
	}

	stockTrans := stockAmount
	moneyTrans := sdk.NewDecFromInt(stockAmount).Mul(orderInBook.ActualPrice()).TruncateInt()
	if currOrder.IsBuy {
		dealInfo.RemainAmount.Sub(moneyTrans)
	} else {
		dealInfo.RemainAmount.Sub(stockTrans)
	}

	dealInfo.DealMoneyInBook.Add(moneyTrans)
	dealInfo.DealStockInBook.Add(stockTrans)
	// transfer tokens in orders
	if currOrder.IsBuy {
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Money(), moneyTrans)
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Stock(), stockTrans)
	} else {
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Stock(), stockTrans)
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Money(), moneyTrans)
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
		stockAmount = sdk.NewDecFromInt(dealInfo.RemainAmount).Quo(order.ActualPrice()).TruncateInt()
	} else {
		stockAmount = dealInfo.RemainAmount
	}
	if stockAmount.IsPositive() {
		order.Amount = stockAmount.Int64()
		if dealInfo.HasDealInOrderBook {
			pk.insertOrderAtHead(ctx, order)
		} else {
			pk.insertOrderFromHead(ctx, order)
		}
	}
	if order.IsBuy {
		poolInfo.moneyOrderBookReserve.Add(moneyAmount)
	} else {
		poolInfo.stockOrderBookReserve.Add(stockAmount)
	}
}

func (pk PairKeeper) insertOrderAtHead(ctx sdk.Context, order *types.Order) {
	order.NextOrderID = pk.GetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy)
	pk.SetOrder(ctx, order)
	pk.SetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy, order.OrderID)
}

func (pk PairKeeper) insertOrderFromHead(ctx sdk.Context, order *types.Order) bool {
	firstOrderID := pk.GetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy)
	var (
		firstOrder *types.Order
		canBeFirst = firstOrderID <= 0
	)
	if !canBeFirst {
		firstOrder = pk.GetOrder(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy, firstOrderID)
		canBeFirst = (order.IsBuy && order.Price > firstOrder.Price) ||
			(!order.IsBuy && order.Price < firstOrder.Price)
	}
	if canBeFirst {
		order.NextOrderID = firstOrderID
		pk.SetOrder(ctx, order)
		pk.SetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy, order.OrderID)
		return true
	}
	return pk.insertOrder(ctx, order, firstOrder)
}

func (pk PairKeeper) dealWithPoolAndCollectFee(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) sdk.Int {
	outPoolTokenReserve, inPoolTokenReserve, otherToTaker := poolInfo.moneyAmmReserve, poolInfo.stockAmmReserve, dealInfo.DealMoneyInBook
	if order.IsBuy {
		outPoolTokenReserve, inPoolTokenReserve, otherToTaker = poolInfo.stockAmmReserve, poolInfo.moneyAmmReserve, dealInfo.DealStockInBook
	}
	outAmount := outPoolTokenReserve.Mul(dealInfo.AmountInToPool).Quo(inPoolTokenReserve.Add(dealInfo.AmountInToPool))
	if dealInfo.AmountInToPool.IsPositive() {
		// todo emit deal with pool log
	}

	amountToTaker := outAmount.Add(otherToTaker)
	// todo. will add fee calculate and setup fee param in param store
	fee := sdk.ZeroInt()
	amountToTaker = amountToTaker.Sub(fee)
	if order.IsBuy {
		poolInfo.moneyAmmReserve = poolInfo.moneyAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.stockAmmReserve = poolInfo.stockAmmReserve.Sub(outAmount).Add(fee)
	} else {
		poolInfo.stockAmmReserve = poolInfo.stockAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.moneyAmmReserve = poolInfo.moneyAmmReserve.Sub(outPoolTokenReserve).Add(fee)
	}
	// todo. transfer token from pool to order sender
	moduleAccount, _ := sdk.AccAddressFromHex("moduleAccount")
	if order.IsBuy {
		pk.transferToken(ctx, moduleAccount, order.Sender, order.Stock(), outAmount)
		pk.transferToken(ctx, order.Sender, moduleAccount, order.Money(), dealInfo.AmountInToPool)
	} else {
		pk.transferToken(ctx, moduleAccount, order.Sender, order.Money(), outAmount)
		pk.transferToken(ctx, order.Sender, moduleAccount, order.Stock(), dealInfo.AmountInToPool)
	}
	return amountToTaker
}

func (pk PairKeeper) AddMarketOrder(ctx sdk.Context, order *types.Order) error {
	order.Price = 0
	if order.IsBuy {
		order.Price = math.MaxInt64
	}
	poolInfo := pk.GetPoolInfo(ctx, order.MarketSymbol, order.IsOpenSwap)
	if _, err := pk.dealOrderAndAddRemainedOrder(ctx, order, poolInfo); err != nil {
		return err
	}
	return nil
}

func (pk PairKeeper) HasOrder(ctx sdk.Context, symbol string, isOpenSwap, isBuy bool, orderID int64) bool {
	return pk.GetOrder(ctx, symbol, isOpenSwap, isBuy, orderID) != nil
}

func (pk PairKeeper) GetOrder(ctx sdk.Context, symbol string, isOpenSwap bool, isBuy bool, orderID int64) (order *types.Order) {
	store := ctx.KVStore(pk.storeKey)
	key := getOrderKey(&types.Order{
		OrderBasic: types.OrderBasic{
			MarketSymbol: symbol,
			IsOpenSwap:   isOpenSwap,
			IsBuy:        isBuy,
		},
		OrderID: orderID,
	})
	val := store.Get(key)
	if len(val) == 0 {
		return nil
	}
	pk.codec.MustUnmarshalBinaryBare(val, order)
	return order
}

func (pk PairKeeper) SetOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(pk.storeKey)
	key := getOrderKey(order)
	val := pk.codec.MustMarshalBinaryBare(order)
	store.Set(key, val)
}

func (pk PairKeeper) DeleteOrder(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(pk.storeKey)
	key := getOrderKey(order)
	store.Delete(key)
}

func (pk PairKeeper) GetFirstOrderID(ctx sdk.Context, symbol string, isOpenSwap, isBuy bool) int64 {
	store := ctx.KVStore(pk.storeKey)
	key := getBestOrderPriceKey(symbol, isOpenSwap, isBuy)
	val := store.Get(key)
	if len(val) == 0 {
		return -1
	}
	return int64(binary.BigEndian.Uint64(val))
}

func (pk PairKeeper) SetFirstOrderID(ctx sdk.Context, symbol string, isOpenSwap, isBuy bool, orderID int64) {
	store := ctx.KVStore(pk.storeKey)
	key := getBestOrderPriceKey(symbol, isOpenSwap, isBuy)
	val := strconv.Itoa(int(orderID))
	store.Set(key, []byte(val))
}
