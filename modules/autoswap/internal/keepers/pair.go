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

func (pk PairKeeper) AddLimitOrder(ctx sdk.Context, order *types.Order) error {
	var poolInfo *PoolInfo
	if poolInfo = pk.GetPoolInfo(ctx, order.MarketSymbol, order.IsOpenSwap); poolInfo == nil {
		return fmt.Errorf("can't find pool info")
	}
	order.OrderID = pk.getUnusedOrderID(ctx, order)
	if order.OrderID <= 0 {
		return fmt.Errorf("can't find available order id")
	}
	//1. todo. will calculate and check order amount

	//2. calculate insert order position, try insert order if the order can't deal.
	if order.HasPrevKey() {
		if pk.insertOrderFromGivePos(ctx, order) {
			//3. when insert order later, update pool info
			if order.IsBuy {
				poolInfo.moneyOrderBookReserve.Add(sdk.Int{})
			} else {
				poolInfo.stockOrderBookReserve.Add(sdk.Int{})
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
func (pk PairKeeper) dealOrderAndAddRemainedOrder(ctx sdk.Context, order *types.Order, poolInfo *PoolInfo) (int64, error) {
	firstOrderID := pk.GetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy)
	currOrderID := firstOrderID
	dealInfo := &types.DealInfo{RemainAmount: order.Amount}
	if order.IsBuy {
		dealInfo.RemainAmount = order.Amount * order.Price
	}

	for currOrderID > 0 {
		orderInBook := pk.GetOrder(ctx, order.MarketSymbol, order.IsOpenSwap, !order.IsBuy, currOrderID)
		canDealInBook := (order.IsBuy && order.Price >= orderInBook.Price) ||
			(!order.IsBuy && order.Price <= orderInBook.Price)
		// can't deal with order book
		if !canDealInBook {
			break
		}
		// full deal in pool
		if allDeal := pk.tryDealInPool(dealInfo, orderInBook.Price, order.IsBuy); allDeal {
			break
		}
		// deal in order book
		if err := pk.dealInOrderBook(ctx, order, orderInBook, dealInfo); err != nil {
			return 0, err
		}
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
		pk.tryDealInPool(dealInfo, order.Price, order.IsBuy)
		pk.insertOrderToBook(ctx, order, dealInfo, poolInfo)
	} else {
		dealInfo.AmountInToPool += order.Amount
		dealInfo.RemainAmount = 0
	}
	amountToTaker := pk.dealWithPoolAndCollectFee(ctx, order, dealInfo, poolInfo)
	poolInfo := pk.GetPoolInfo(ctx, order.MarketSymbol, order.IsOpenSwap)
	if order.IsBuy {
		poolInfo.stockOrderBookReserve -= dealInfo.DealStockInBook
	} else {
		poolInfo.moneyOrderBookReserve -= dealInfo.DealMoneyInBook
	}
	if dealInfo.AmountInToPool > 0 {
		if order.IsBuy {
			poolInfo.moneyAmmReserve += dealInfo.AmountInToPool * order.Price
			poolInfo.stockAmmReserve -= dealInfo.AmountInToPool
		} else {
			poolInfo.stockAmmReserve += dealInfo.AmountInToPool
			poolInfo.moneyAmmReserve -= dealInfo.AmountInToPool * order.Price
		}
	}
	if firstOrderID != currOrderID {
		pk.SetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, !order.IsBuy, currOrderID)
	}
	pk.SetPoolInfo(ctx, order.MarketSymbol, order.IsOpenSwap, poolInfo)
	return amountToTaker, nil
}

func (pk PairKeeper) tryDealInPool(dealInfo *types.DealInfo, dealPrice uint64, isBuy bool) bool {
	currTokenCanTradeWithPool := pk.intoPoolAmountTillPrice(dealPrice, isBuy)
	// todo. will check deal token amount later.

	// todo. round amount
	if !isBuy {
	}
	if currTokenCanTradeWithPool > dealInfo.AmountInToPool {
		diffTokenTradeWithPool := currTokenCanTradeWithPool - dealInfo.AmountInToPool
		allDeal := diffTokenTradeWithPool > dealInfo.RemainAmount
		if allDeal {
			diffTokenTradeWithPool = dealInfo.RemainAmount
		}
		dealInfo.AmountInToPool += diffTokenTradeWithPool
		dealInfo.RemainAmount -= diffTokenTradeWithPool
		return allDeal
	}
	return false
}

func (pk PairKeeper) intoPoolAmountTillPrice(dealPrice uint64, isBuy bool) uint64 {
	return 0
}

func (pk PairKeeper) dealInOrderBook(ctx sdk.Context, currOrder, orderInBook *types.Order, dealInfo *types.DealInfo) error {
	dealInfo.HasDealInOrderBook = true
	stockAmount := 0
	// todo. will calculate stock amount
	if currOrder.IsBuy {
		//dealInfo.RemainAmount
		stockAmount = 1
	} else {
		stockAmount = 2
	}

	if orderInBook.Amount < stockAmount {
		stockAmount = orderInBook.Amount
	}
	// todo. check stock amount

	stockTrans := stockAmount
	moneyTrans := stockAmount * currOrder.Price
	orderInBook.Amount -= stockAmount
	if currOrder.IsBuy {
		dealInfo.RemainAmount -= moneyTrans
	} else {
		dealInfo.RemainAmount -= stockTrans
	}

	dealInfo.DealMoneyInBook += moneyTrans
	dealInfo.DealStockInBook += stockTrans
	if currOrder.IsBuy {
		return pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, moneyTrans)
	} else {
		return pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, stockTrans)
	}
}

func (pk PairKeeper) transferToken(ctx sdk.Context, from, to sdk.AccAddress, token string, amount sdk.Int) error {
	coin := newCoins(token, amount)
	if err := pk.UnFreezeCoins(ctx, from, coin); err != nil {
		return err
	}
	if err := pk.SendCoins(ctx, from, to, coin); err != nil {
		return err
	}
	return nil
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
		stockAmount = dealInfo.RemainAmount / order.Price
	} else {
		stockAmount = dealInfo.RemainAmount
	}
	if stockAmount > 0 {
		order.Amount = stockAmount
		if dealInfo.HasDealInOrderBook {
			pk.insertOrderAtHead(ctx, order)
		} else {
			pk.insertOrderFromHead(ctx, order)
		}
	}
	if order.IsBuy {
		poolInfo.moneyOrderBookReserve += moneyAmount
	} else {
		poolInfo.stockOrderBookReserve += stockAmount
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

func (pk PairKeeper) dealWithPoolAndCollectFee(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) int64 {
	outPoolTokenReserve, inPoolTokenReserve, otherToTaker := poolInfo.moneyAmmReserve, poolInfo.stockAmmReserve, dealInfo.DealMoneyInBook
	if order.IsBuy {
		outPoolTokenReserve, inPoolTokenReserve, otherToTaker := poolInfo.stockAmmReserve, poolInfo.moneyAmmReserve, dealInfo.DealStockInBook
	}
	outAmount := (outPoolTokenReserve * dealInfo.AmountInToPool) / (inPoolTokenReserve + dealInfo.AmountInToPool)
	if dealInfo.AmountInToPool > 0 {
		// emit log todo.
	}
	// todo. will add fee calculate and setup fee param in param store
	amountToTaker := outAmount + otherToTaker
	fee := 0
	amountToTaker -= fee
	if order.IsBuy {
		poolInfo.moneyAmmReserve += dealInfo.AmountInToPool
		poolInfo.stockAmmReserve = poolInfo.stockAmmReserve - outAmount + fee
	} else {
		poolInfo.stockAmmReserve += dealInfo.AmountInToPool
		poolInfo.moneyAmmReserve = poolInfo.moneyAmmReserve - outPoolTokenReserve + fee
	}
	token := order.Money()
	if order.IsBuy {
		token = order.Stock()
	}
	_ = pk.transferToken(ctx, moduleAccount, order.Sender, token, amountToTaker)
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
