package keepers

import (
	"encoding/binary"
	"fmt"

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
	codec    *codec.Codec
	storeKey sdk.StoreKey
}

func NewPairKeeper(poolKeeper IPoolKeeper, codec *codec.Codec, storeKey sdk.StoreKey) *PairKeeper {
	return &PairKeeper{
		codec:       codec,
		storeKey:    storeKey,
		IPoolKeeper: poolKeeper,
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
	pk.addOrder(ctx, order)
	return nil
}

func (pk PairKeeper) getUnusedOrderID(ctx sdk.Context, order *types.Order) int64 {
	var id int64 = 0
	if order.OrderID == 0 {
		id = int64(binary.BigEndian.Uint64(ctx.BlockHeader().AppHash[:8]) ^
			binary.BigEndian.Uint64(order.Sender[:8]))
	}

	for i := 0; i < 100 && id > 0; i++ {
		if pk.HasOrder(ctx, order.IsBuy, id) {
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

// addOrder Deal the order and
// insert the remainder order into the order book
func (pk PairKeeper) addOrder(ctx sdk.Context, order *types.Order) {
	firstOrderID := pk.GetFirstOrderID(ctx, order.MarketSymbol, order.IsOpenSwap, order.IsBuy)
	currOrderID := firstOrderID
	dealInfo := &types.DealInfo{}
	for currOrderID > 0 {
		orderInBook := pk.GetOrder(ctx, order.MarketSymbol, order.IsOpenSwap, !order.IsBuy, currOrderID)
		canDealInBook := (order.IsBuy && order.Price >= orderInBook.Price) ||
			(!order.IsBuy && order.Price <= orderInBook.Price)
		if !canDealInBook {
			break
		}
		pk.tryDealInPool(ctx, dealInfo, orderInBook.Price, order.IsBuy)
	}
}

func (pk PairKeeper) tryDealInPool(ctx sdk.Context, dealInfo *types.DealInfo, dealPrice uint64, isBuy bool) bool {
	currTokenCanTradeWithPool := pk.intoPoolAmountTillPrice(dealPrice, isBuy)
	// todo. will check deal token amount later.

	// todo. round amount
	if !isBuy {
	}

	if currTokenCanTradeWithPool > dealInfo.DealWithPoolAmount {
		diffTokenTradeWithPool := currTokenCanTradeWithPool - dealInfo.DealWithPoolAmount
		allDeal := diffTokenTradeWithPool > dealInfo.RemainAmount
		if allDeal {
			diffTokenTradeWithPool = dealInfo.RemainAmount
		}
		dealInfo.DealWithPoolAmount += diffTokenTradeWithPool
		dealInfo.RemainAmount -= diffTokenTradeWithPool
		return allDeal
	}
	return false
}

func (pk PairKeeper) intoPoolAmountTillPrice(dealPrice uint64, isBuy bool) uint64 {
	return 0
}

func (pk PairKeeper) AddMarketOrder(ctx sdk.Context, order *types.Order) bool {
	panic("implement me")
}

func (pk PairKeeper) HasOrder(ctx sdk.Context, isBuy bool, orderID int64) bool {
	panic("implement me")
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

func (pk PairKeeper) GetFirstOrderID(ctx sdk.Context, symbol string, isOpenSwap, isBuy bool) int64 {
	store := ctx.KVStore(pk.storeKey)
	key := getBestOrderPriceKey(symbol, isOpenSwap, isBuy)
	val := store.Get(key)
	if len(val) == 0 {
		return -1
	}
	return int64(binary.BigEndian.Uint64(val))
}
