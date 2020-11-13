package keepers

import (
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

//type IOrderBookKeeper interface {
//	AddOrder(sdk.Context, *types.Order)
//	DelOrder(sdk.Context, *types.Order)
//	GetOrder(sdk.Context, *QueryOrderInfo) *types.Order
//	GetBestPrice(ctx sdk.Context, market string, isBuy bool) sdk.Dec
//	GetMatchedOrder(ctx sdk.Context, order *types.Order) []*types.Order
//	OrderIndexInOneBlock() int32
//	ResetOrderIndexInOneBlock()
//}

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
