package keepers

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"

	sdkstore "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func newContextAndStoreKey(t *testing.T) (sdk.Context, sdk.StoreKey) {
	db := dbm.NewMemDB()
	ms := sdkstore.NewCommitMultiStore(db)
	key := sdk.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	require.NoError(t, ms.LoadLatestVersion())
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	return ctx, key
}

func TestOrderIndexInOneBlock(t *testing.T) {
	_, storeKey := newContextAndStoreKey(t)
	orderKeeper := NewOrderKeeper(codec.New(), storeKey)

	orderKeeper.ordersIndexInOneBlock = 9
	orderKeeper.ResetOrderIndexInOneBlock()
	require.EqualValues(t, 0, orderKeeper.ordersIndexInOneBlock)
}

func TestOrderKeeper_AddOrder_GetOrder_DelOrder(t *testing.T) {
	ctx, storeKey := newContextAndStoreKey(t)
	orderKeeper := NewOrderKeeper(codec.New(), storeKey)
	tradingPair := "abc/def"
	baseOrder := types.Order{
		TradingPair: tradingPair,
		Sender:      supply.NewModuleAddress("aass"),
		Sequence:    1,
		Identify:    1,
		Price:       sdk.NewDec(10),
		Quantity:    1000,
		Height:      1,
		IsBuy:       true,
		LeftStock:   1000,
		Freeze:      1000000,
		DealMoney:   10,
		DealStock:   20,
	}

	// add orders
	orders := make([]types.Order, 0, 4)
	orderIDs := make([]string, 0, 4)
	for i := int64(1); i < 5; i++ {
		order := baseOrder
		order.Sequence = i + 1
		order.Price = sdk.NewDec(i * 10)
		order.OrderIndexInOneBlock = int32(i - 1)
		orders = append(orders, order)
		orderIDs = append(orderIDs, order.GetOrderID())
		orderKeeper.AddOrder(ctx, &order)
	}
	// check index in keeper
	require.EqualValues(t, 4, orderKeeper.ordersIndexInOneBlock)

	queryOrders := orderKeeper.GetAllOrdersInMarket(ctx, tradingPair)
	for _, v := range queryOrders {
		fmt.Println(v.GetOrderID())
	}
	require.EqualValues(t, 4, len(queryOrders))
	require.EqualValues(t, 4, len(orderKeeper.GetOrdersFromUser(ctx, supply.NewModuleAddress("aass").String())))

	// get orders
	for i, id := range orderIDs {
		queryOrder := orderKeeper.GetOrder(ctx, &QueryOrderInfo{OrderID: id})
		require.EqualValues(t, orders[i], *queryOrder)
	}

	// del orders [0, 1, 2, 3]
	orderKeeper.DelOrder(ctx, &orders[1])
	require.Nil(t, orderKeeper.GetOrder(ctx, &QueryOrderInfo{OrderID: orderIDs[1]}))
	orderKeeper.DelOrder(ctx, &orders[3])
	require.Nil(t, orderKeeper.GetOrder(ctx, &QueryOrderInfo{OrderID: orderIDs[3]}))
	require.EqualValues(t, 2, len(orderKeeper.GetOrdersFromUser(ctx, supply.NewModuleAddress("aass").String())))
}

func TestOrderKeeper_GetBestSellPrice(t *testing.T) {
	ctx, storeKey := newContextAndStoreKey(t)
	orderKeeper := NewOrderKeeper(codec.New(), storeKey)
	tradingPair := "abc/def"
	baseOrder := types.Order{
		TradingPair: tradingPair,
		Sender:      supply.NewModuleAddress("aass"),
		Sequence:    1,
		Identify:    1,
		Price:       sdk.NewDec(10),
		Quantity:    1000,
		Height:      1,
		IsBuy:       true,
		LeftStock:   1000,
		Freeze:      1000000,
		DealMoney:   10,
		DealStock:   20,
	}

	// check sell orders
	baseOrder.IsBuy = false
	sellOrders := make([]types.Order, 0, 4)
	sellOrderIDs := make([]string, 0, 4)
	for i := int64(1); i < 7; i++ {
		order := baseOrder
		order.Sequence = i + 10
		order.Identify = byte(i)
		if i%2 == 0 {
			order.Price = sdk.NewDec(i * 10)
		} else {
			order.Price = sdk.NewDec(i * 30)
		}
		order.OrderIndexInOneBlock = int32(i - 1)
		sellOrders = append(sellOrders, order)
		sellOrderIDs = append(sellOrderIDs, order.GetOrderID())
		orderKeeper.AddOrder(ctx, &order)
	}
	// sell order prices: [30, 20, 90, 40, 150, 60]
	bestOrder := orderKeeper.GetOrder(ctx, &QueryOrderInfo{OrderID: sellOrderIDs[1]})
	bestPrice := orderKeeper.GetBestPrice(ctx, tradingPair, false)
	require.EqualValues(t, sdk.NewDec(20), bestPrice)
	require.EqualValues(t, sdk.NewDec(20), bestOrder.Price)

	orderKeeper.DelOrder(ctx, &sellOrders[1])
	bestOrder = orderKeeper.GetOrder(ctx, &QueryOrderInfo{OrderID: sellOrderIDs[0]})
	bestPrice = orderKeeper.GetBestPrice(ctx, tradingPair, false)
	require.EqualValues(t, sdk.NewDec(30), bestPrice)
	require.EqualValues(t, sdk.NewDec(30), bestOrder.Price)

	orderKeeper.DelOrder(ctx, &sellOrders[0])
	bestPrice = orderKeeper.GetBestPrice(ctx, tradingPair, false)
	require.EqualValues(t, sdk.NewDec(40), bestPrice)

	orderKeeper.DelOrder(ctx, &sellOrders[3])
	bestPrice = orderKeeper.GetBestPrice(ctx, tradingPair, false)
	require.EqualValues(t, sdk.NewDec(60), bestPrice)
}

func TestOrderKeeper_GetBestBuyPrice(t *testing.T) {
	ctx, storeKey := newContextAndStoreKey(t)
	orderKeeper := NewOrderKeeper(codec.New(), storeKey)
	tradingPair := "abc/def"
	baseOrder := types.Order{
		TradingPair: tradingPair,
		Sender:      supply.NewModuleAddress("aass"),
		Sequence:    1,
		Identify:    1,
		Price:       sdk.NewDec(10),
		Quantity:    1000,
		Height:      1,
		IsBuy:       true,
		LeftStock:   1000,
		Freeze:      1000000,
		DealMoney:   10,
		DealStock:   20,
	}

	//check buyOrders
	baseOrder.Sender = supply.NewModuleAddress("dffwie")
	buyOrders := make([]types.Order, 0, 4)
	buyOrderIDs := make([]string, 0, 4)
	for i := int64(1); i < 7; i++ {
		order := baseOrder
		order.Sequence = i + 1
		if i%2 == 0 {
			order.Price = sdk.NewDec(i * 10)
		} else {
			order.Price = sdk.NewDec(i * 30)
		}
		order.OrderIndexInOneBlock = int32(i - 1)
		buyOrders = append(buyOrders, order)
		buyOrderIDs = append(buyOrderIDs, order.GetOrderID())
		orderKeeper.AddOrder(ctx, &order)
	}

	// buy order prices: [30, 20, 90, 40, 150, 60]
	bestOrder := orderKeeper.GetOrder(ctx, &QueryOrderInfo{OrderID: buyOrderIDs[4]})
	bestPrice := orderKeeper.GetBestPrice(ctx, tradingPair, true)
	require.EqualValues(t, sdk.NewDec(150), bestPrice)
	require.EqualValues(t, sdk.NewDec(150), bestOrder.Price)

	orderKeeper.DelOrder(ctx, &buyOrders[4])
	bestPrice = orderKeeper.GetBestPrice(ctx, tradingPair, true)
	require.EqualValues(t, sdk.NewDec(90), bestPrice)

	orderKeeper.DelOrder(ctx, &buyOrders[2])
	bestPrice = orderKeeper.GetBestPrice(ctx, tradingPair, true)
	require.EqualValues(t, sdk.NewDec(60), bestPrice)
}

func TestOrderKeeper_GetFirstSamePriceIndex(t *testing.T) {
	orders := make([]*types.Order, 0, 4)
	for i := int64(0); i < 4; i++ {
		orders = append(orders, &types.Order{Price: sdk.NewDec((i + 1) * 10)})
	}
	//prices: [10,20,30,40]
	require.EqualValues(t, 3, getFirstSamePriceIndex(orders))

	//prices: [10,20,40,40]
	orders[2] = &types.Order{Price: sdk.NewDec(40)}
	require.EqualValues(t, 2, getFirstSamePriceIndex(orders))

	//prices: [10,20,40,40]
	orders[1] = &types.Order{Price: sdk.NewDec(40)}
	require.EqualValues(t, 1, getFirstSamePriceIndex(orders))

	//prices: [40,40,40,40]
	orders[0] = &types.Order{Price: sdk.NewDec(40)}
	require.EqualValues(t, 0, getFirstSamePriceIndex(orders))
}

func TestOrderKeeper_sortOrderWithCreateHeightAndTxIndex(t *testing.T) {
	orders := make([]*types.Order, 0, 4)
	for i := int64(0); i < 4; i++ {
		orders = append(orders, &types.Order{Height: i})
	}
	// heights: [0,1,2,3]
	sortOrder := sortOrderWithCreateHeightAndTxIndex(orders)
	require.EqualValues(t, orders, sortOrder)

	// heights: [0,1,3,3]; indexes: [0,0,1,0]
	orders[2] = &types.Order{OrderIndexInOneBlock: 1, Height: 3}
	sortOrder = sortOrderWithCreateHeightAndTxIndex(orders)
	orders[2], orders[3] = orders[3], orders[2]
	require.EqualValues(t, orders, sortOrder)
}

func TestOrderKeeper_GetMatchedOrder(t *testing.T) {
	ctx, storeKey := newContextAndStoreKey(t)
	orderKeeper := NewOrderKeeper(codec.New(), storeKey)
	tradingPair := "abc/def"
	baseOrder := types.Order{
		TradingPair: tradingPair,
		Sender:      supply.NewModuleAddress("buy"),
		Sequence:    1,
		Identify:    1,
		Price:       sdk.NewDec(10),
		Quantity:    1000,
		Height:      1,
		IsBuy:       true,
		Freeze:      1000000,
		DealMoney:   10,
		DealStock:   20,
	}

	// one block to fill buy orders
	buyOrder := baseOrder
	buyOrders := make([]*types.Order, 0, 4)
	buyOrderIDs := make([]string, 0, 4)
	// prices: [10, 30, 30, 90, 50, 150]
	// amounts: [1000, 3000, 3000, 9000, 5000, 15000]
	for i := int64(0); i < 6; i++ {
		order := buyOrder
		order.Sequence = i + 1
		if i%2 == 0 {
			order.Price = sdk.NewDec((i + 1) * 10)
			order.LeftStock = (i + 1) * 1000
		} else {
			order.Price = sdk.NewDec(i * 30)
			order.LeftStock = i * 3000
		}
		order.OrderIndexInOneBlock = int32(i)
		buyOrders = append(buyOrders, &order)
		buyOrderIDs = append(buyOrderIDs, order.GetOrderID())
		orderKeeper.AddOrder(ctx, &order)
		fmt.Println("buy order: ", order.GetOrderID())
	}

	// one block to fill sell orders
	sellOrder := baseOrder
	sellOrder.IsBuy = false
	sellOrder.Sender = supply.NewModuleAddress("sell")
	sellOrders := make([]*types.Order, 0, 4)
	sellOrderIDs := make([]string, 0, 4)
	// prices: [20, 20, 40, 60, 60, 100]
	// amounts: [2000, 2000, 4000, 6000, 6000, 10000]
	for i := int64(0); i < 6; i++ {
		order := sellOrder
		order.Sequence = i + 1
		if i%2 == 0 {
			order.Price = sdk.NewDec((i + 2) * 10)
			order.LeftStock = (i + 2) * 1000
		} else {
			order.Price = sdk.NewDec(i * 20)
			order.LeftStock = i * 2000
		}
		order.OrderIndexInOneBlock = int32(i)
		sellOrders = append(sellOrders, &order)
		sellOrderIDs = append(sellOrderIDs, order.GetOrderID())
		orderKeeper.AddOrder(ctx, &order)
		fmt.Println("sell order: ", order.GetOrderID())
	}

	// buy order prices: [10, 30, 30, 90, 50, 150], amounts = [1000, 3000, 3000, 9000, 5000, 15000]
	// sell order prices: [20, 20, 40, 60, 60, 100], amounts = [2000, 2000, 4000, 6000, 6000, 10000]
	require.EqualValues(t, 0, len(orderKeeper.GetMatchedOrder(ctx, buyOrders[0])), "should not have matched order")
	require.EqualValues(t, 2, len(orderKeeper.GetMatchedOrder(ctx, buyOrders[1])), "should have 2 matched order")
	require.EqualValues(t, 1, len(orderKeeper.GetMatchedOrder(ctx, sellOrders[2])), "should have 1 matched order")
	require.EqualValues(t, 1, len(orderKeeper.GetMatchedOrder(ctx, sellOrders[4])), "should have 1 matched order")

	// modify sell order quantity
	sellOrders[4].LeftStock = 16000
	orderKeeper.AddOrder(ctx, sellOrders[4])
	require.EqualValues(t, 2, len(orderKeeper.GetMatchedOrder(ctx, sellOrders[4])), "should have 1 matched order")
}
