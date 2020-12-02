package keepers

import (
	"bytes"
	"sort"

	"github.com/coinexchain/cet-sdk/modules/market"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ IOrderBookKeeper = &OrderKeeper{}

type IOrderBookKeeper interface {
	AddOrder(sdk.Context, *types.Order)
	DelOrder(sdk.Context, *types.Order)
	GetAllOrdersInMarket(ctx sdk.Context, market string) []*types.Order
	GetOrdersFromUser(ctx sdk.Context, user string) []string
	StoreToOrderBook(ctx sdk.Context, order *types.Order)
	GetOrder(sdk.Context, *market.QueryOrderParam) *types.Order
	GetBestPrice(ctx sdk.Context, market string, isBuy bool) sdk.Dec
	GetMatchedOrder(ctx sdk.Context, order *types.Order) []*types.Order
	OrderIndexInOneBlock() int32
	ResetOrderIndexInOneBlock()
}

type OrderKeeper struct {
	codec                 *codec.Codec
	storeKey              sdk.StoreKey
	ordersIndexInOneBlock int32
}

func NewOrderKeeper(codec *codec.Codec, storeKey sdk.StoreKey) *OrderKeeper {
	return &OrderKeeper{codec: codec, storeKey: storeKey}
}

func (o *OrderKeeper) OrderIndexInOneBlock() int32 {
	index := o.ordersIndexInOneBlock
	o.ordersIndexInOneBlock++
	return index
}

func (o *OrderKeeper) ResetOrderIndexInOneBlock() {
	o.ordersIndexInOneBlock = 0
}

func (o *OrderKeeper) AddOrder(ctx sdk.Context, order *types.Order) {
	o.storeBidOrAskQueue(ctx, order)
	o.StoreToOrderBook(ctx, order)
	o.storeOrderToMarket(ctx, order)
}

func (o *OrderKeeper) storeBidOrAskQueue(ctx sdk.Context, order *types.Order) {
	var (
		sideKey []byte
		store   = ctx.KVStore(o.storeKey)
	)
	order.OrderIndexInOneBlock = o.OrderIndexInOneBlock()
	if order.IsBuy {
		sideKey = getBidOrderKey(order)
	} else {
		sideKey = getAskOrderKey(order)
	}
	store.Set(sideKey, []byte{0x0})
}

func (o *OrderKeeper) StoreToOrderBook(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(o.storeKey)
	key := getOrderBookKey(order.GetOrderID())
	val := o.codec.MustMarshalBinaryBare(order)
	store.Set(key, val)
}

func (o *OrderKeeper) storeOrderToMarket(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(o.storeKey)
	key := getOrderKeyInMarket(order.TradingPair, order.GetOrderID())
	store.Set(key, []byte{0x0})
}

func (o *OrderKeeper) DelOrder(ctx sdk.Context, order *types.Order) {
	o.delOrderBook(ctx, order)
	o.delBidOrAskQueue(ctx, order)
}

func (o *OrderKeeper) delOrderBook(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(o.storeKey)
	key := getOrderBookKey(order.GetOrderID())
	store.Delete(key)
}

func (o *OrderKeeper) delBidOrAskQueue(ctx sdk.Context, order *types.Order) {
	var (
		sideKey []byte
		store   = ctx.KVStore(o.storeKey)
	)
	if order.IsBuy {
		sideKey = getBidOrderKey(order)
	} else {
		sideKey = getAskOrderKey(order)
	}
	store.Delete(sideKey)
}

func (o *OrderKeeper) delOrderInMarket(ctx sdk.Context, order *types.Order) {
	store := ctx.KVStore(o.storeKey)
	key := getOrderKeyInMarket(order.TradingPair, order.GetOrderID())
	store.Delete(key)
}

func (o OrderKeeper) GetOrder(ctx sdk.Context, info *market.QueryOrderParam) *types.Order {
	var (
		order = types.Order{}
		store = ctx.KVStore(o.storeKey)
	)
	key := getOrderBookKey(info.OrderID)
	val := store.Get(key)
	if len(val) == 0 {
		return nil
	}
	o.codec.MustUnmarshalBinaryBare(val, &order)
	return &order
}

func (o OrderKeeper) GetBestPrice(ctx sdk.Context, tradingPair string, isBuy bool) sdk.Dec {
	var (
		key   []byte
		iter  sdk.Iterator
		store = ctx.KVStore(o.storeKey)
	)
	begin, end := getBidOrAskQueueBeginEndKey(tradingPair, isBuy)
	if isBuy {
		iter = store.ReverseIterator(begin, end)
	} else {
		iter = store.Iterator(begin, end)
	}
	defer iter.Close()
	if iter.Valid() {
		key = iter.Key()
	} else {
		panic("invalid iterator")
	}
	pos := getOrderIDPos(tradingPair)
	if order := o.GetOrder(ctx, &market.QueryOrderParam{OrderID: string(key[pos:])}); order != nil {
		return order.Price
	}
	panic(types.ErrInvalidOrderID(string(key[pos:])))
}

func getBidOrAskQueueBeginEndKey(tradingPair string, isBuy bool) ([]byte, []byte) {
	var (
		begin []byte
		end   []byte
	)
	if isBuy {
		begin = getBidQueueBegin(tradingPair)
		end = getBidQueueEnd(tradingPair)
	} else {
		begin = getAskQueueBegin(tradingPair)
		end = getAskQueueEnd(tradingPair)
	}
	return begin, end
}

func (o OrderKeeper) GetMatchedOrder(ctx sdk.Context, order *types.Order) []*types.Order {
	var (
		key        []byte
		orderIDPos int
		pricePoses []int
		iter       sdk.Iterator
		store      = ctx.KVStore(o.storeKey)
	)
	begin, end := getBidOrAskQueueBeginEndKey(order.TradingPair, !order.IsBuy)
	if !order.IsBuy {
		iter = store.ReverseIterator(begin, end)
	} else {
		iter = store.Iterator(begin, end)
	}
	defer iter.Close()
	orderPriceBytes := market.DecToBigEndianBytes(order.Price)

	// key = prefix | tradingPair | side | 0x0 | price | orderIndexInOneBlock | orderID
	oppositeOrderIDs := make([]string, 0)
	for ; iter.Valid(); iter.Next() {
		key = iter.Key()
		pricePoses = getPricePos(order.TradingPair)
		if (order.IsBuy && bytes.Compare(orderPriceBytes, key[pricePoses[0]:pricePoses[1]]) >= 0) ||
			(!order.IsBuy && bytes.Compare(orderPriceBytes, key[pricePoses[0]:pricePoses[1]]) <= 0) {
			orderIDPos = getOrderIDPos(order.TradingPair)
			oppositeOrderIDs = append(oppositeOrderIDs, string(key[orderIDPos:]))
		}
	}

	oppositeOrders := make([]*types.Order, 0, len(oppositeOrderIDs))
	totalAmount := int64(0)
	for _, id := range oppositeOrderIDs {
		opOrder := o.getOrder(store, id)
		if totalAmount < order.LeftStock {
			oppositeOrders = append(oppositeOrders, opOrder)
			totalAmount += opOrder.LeftStock
		} else {
			if opOrder.Price.Equal(oppositeOrders[len(oppositeOrders)-1].Price) {
				oppositeOrders = append(oppositeOrders, opOrder)
			} else {
				break
			}
		}
	}
	return inChronologicalOrders(order, oppositeOrders)
}

func (o OrderKeeper) getOrder(store sdk.KVStore, orderID string) *types.Order {
	order := types.Order{}
	val := store.Get(getOrderBookKey(orderID))
	o.codec.MustUnmarshalBinaryBare(val, &order)
	return &order
}

func inChronologicalOrders(order *types.Order, oppositeOrders []*types.Order) []*types.Order {
	if len(oppositeOrders) == 0 {
		return nil
	}
	totalAmount := int64(0)
	firstSamePriceIndex := getFirstSamePriceIndex(oppositeOrders)
	samePriceOrders := oppositeOrders[firstSamePriceIndex:]
	sortSamePriceOrders := sortOrderWithCreateHeightAndTxIndex(samePriceOrders)

	// statistical other price amount
	for i := 0; i < firstSamePriceIndex; i++ {
		totalAmount += oppositeOrders[i].LeftStock
	}

	ret := oppositeOrders[:firstSamePriceIndex]
	for i := 0; i < len(sortSamePriceOrders); i++ {
		if totalAmount < order.LeftStock {
			ret = append(ret, sortSamePriceOrders[i])
			totalAmount += sortSamePriceOrders[i].LeftStock
		} else {
			break
		}
	}
	return ret
}

func getFirstSamePriceIndex(oppositeOrders []*types.Order) int {
	firstSamePrice := len(oppositeOrders) - 1
	for i := len(oppositeOrders) - 1; i >= 1; i-- {
		currOrder := oppositeOrders[i]
		if !currOrder.Price.Equal(oppositeOrders[i-1].Price) {
			firstSamePrice = i
			break
		} else {
			if i == 1 {
				firstSamePrice = 0
			}
		}

	}
	return firstSamePrice
}

func sortOrderWithCreateHeightAndTxIndex(orders []*types.Order) []*types.Order {
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Height < orders[j].Height ||
			(orders[i].Height == orders[j].Height &&
				orders[i].OrderIndexInOneBlock < orders[j].OrderIndexInOneBlock) {
			return true
		}
		return false
	})
	return orders
}

func (o *OrderKeeper) GetAllOrdersInMarket(ctx sdk.Context, market string) []*types.Order {
	store := ctx.KVStore(o.storeKey)
	begin := getOrderKeyInMarketBegin(market)
	end := getOrderKeyInMarketEnd(market)
	iter := store.Iterator(begin, end)
	defer iter.Close()

	orderIDs := make([]string, 0)
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		orderID := key[getOrderIdPosInMarket(market):]
		orderIDs = append(orderIDs, string(orderID))
	}

	ret := make([]*types.Order, 0, len(orderIDs))
	for _, id := range orderIDs {
		order := o.getOrder(store, id)
		if order == nil {
			panic(types.ErrInvalidOrderID(id))
		}
		ret = append(ret, order)
	}
	return ret
}

func (o *OrderKeeper) GetOrdersFromUser(ctx sdk.Context, user string) []string {
	store := ctx.KVStore(o.storeKey)
	beginKey := append(OrderBookKey, []byte(user+"-")...)
	endKey := append(append(OrderBookKey, []byte(user)...), []byte{0xff}...)
	var result []string
	iter := store.Iterator(beginKey, endKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k := iter.Key()
		result = append(result, string(k[1:]))
	}
	return result
}
