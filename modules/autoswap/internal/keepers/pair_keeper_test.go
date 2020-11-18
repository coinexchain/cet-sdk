package keepers_test

import (
	"fmt"
	"testing"

	"github.com/coinexchain/cet-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	testapp "github.com/coinexchain/cet-sdk/testapp"
)

var (
	sender = testutil.ToAccAddress("sender")
)

func TestPairKeeper_AddLimitOrder(t *testing.T) {
	app := prepareTestApp(t)
	ctx := app.ctx
	app.AutoSwapKeeper.SetParams(ctx, types.DefaultParams())
	k := app.AutoSwapKeeper.IPairKeeper.(*keepers.PairKeeper)

	// 1. add orders
	market := fmt.Sprintf("%s/%s", stockSymbol, moneySymbol)
	beforeMoneyBalance := app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	beforeStockBalance := app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	msgCreateOrder := types.MsgCreateOrder{
		TradingPair: market,
		Price:       100,
		Quantity:    1000,
		Side:        types.BID,
		Identify:    1,
		Sender:      from,
	}
	buyOrderIDs := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		tmpOrder := *msgCreateOrder.GetOrder()
		tmpOrder.Identify = byte(i)
		buyOrderIDs = append(buyOrderIDs, tmpOrder.GetOrderID())
		require.NoError(t, k.AddLimitOrder(ctx, &tmpOrder))
	}

	// 3. check orderIndex
	require.EqualValues(t, 10, k.OrderIndexInOneBlock())

	// 4. check poolInfo
	poolInfo := k.GetPoolInfo(ctx, market)
	require.EqualValues(t, poolInfo.StockAmmReserve.Int64(), 0, "stock pool reserve should be 0")
	require.EqualValues(t, poolInfo.MoneyAmmReserve.Int64(), 0, "stock pool reserve should be 0")
	require.EqualValues(t, int64(10*100*1000), poolInfo.MoneyOrderBookReserve.Int64(), "orderBook money amount isn't correct")
	require.EqualValues(t, 0, poolInfo.StockOrderBookReserve.Int64(), "orderBook stock amount isn't correct")

	// 5. check account balance
	afterMoneyBalance := app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	afterStockBalance := app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	require.EqualValues(t, int64(10*100*1000), beforeMoneyBalance.Sub(afterMoneyBalance).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 0, beforeStockBalance.Sub(afterStockBalance).Int64(), "balance in account isn't correct")

	// check orders exist
	for _, id := range buyOrderIDs {
		require.NotNil(t, k.GetOrder(ctx, id), "order should be exist")
	}

	// -----------

	// 6. add set order
	beforeMoneyBalance = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	beforeStockBalance = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	msgCreateOrder.Side = types.ASK
	msgCreateOrder.Price = 200
	sellOrderIDs := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		tmpOrder := *msgCreateOrder.GetOrder()
		tmpOrder.Identify = byte(i + 20)
		sellOrderIDs = append(sellOrderIDs, tmpOrder.GetOrderID())
		require.NoError(t, k.AddLimitOrder(ctx, &tmpOrder))
	}

	// 7. check orderIndex
	require.EqualValues(t, 21, k.OrderIndexInOneBlock())

	// 8. check poolInfo
	poolInfo = k.GetPoolInfo(ctx, market)
	require.EqualValues(t, poolInfo.StockAmmReserve.Int64(), 0, "stock pool reserve should be 0")
	require.EqualValues(t, poolInfo.MoneyAmmReserve.Int64(), 0, "stock pool reserve should be 0")
	require.EqualValues(t, int64(10*100*1000), poolInfo.MoneyOrderBookReserve.Int64(), "orderBook money amount isn't correct")
	require.EqualValues(t, int64(10*1000), poolInfo.StockOrderBookReserve.Int64(), "orderBook stock amount isn't correct")

	// 9. check account balance
	afterMoneyBalance = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	afterStockBalance = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	require.EqualValues(t, 0, beforeMoneyBalance.Sub(afterMoneyBalance).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 10*1000, beforeStockBalance.Sub(afterStockBalance).Int64(), "balance in account isn't correct")

	// check orders exist
	for _, id := range sellOrderIDs {
		require.NotNil(t, k.GetOrder(ctx, id), "order should be exist")
	}
}

func TestPairKeeper_DealOrders(t *testing.T) {
	app := prepareTestApp(t)
	ctx := app.ctx
	param := types.DefaultParams()
	app.AutoSwapKeeper.SetParams(ctx, param)
	k := app.AutoSwapKeeper

	market := fmt.Sprintf("%s/%s", stockSymbol, moneySymbol)
	beforeMoneyBalanceFrom := app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	beforeStockBalanceFrom := app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	beforeMoneyBalanceTo := app.AccountKeeper.GetAccount(ctx, to).GetCoins().AmountOf(moneySymbol)
	beforeStockBalanceTo := app.AccountKeeper.GetAccount(ctx, to).GetCoins().AmountOf(stockSymbol)
	msgCreateOrder := types.MsgCreateOrder{
		TradingPair: market,
		Price:       100,
		Quantity:    1000,
		Side:        types.BID,
		Identify:    1,
		Sender:      from,
	}

	// -------- full deal

	// 1. add buy, sell order in price: 100; full deal
	buyOrderMsg := msgCreateOrder
	require.NoError(t, k.AddLimitOrder(ctx, buyOrderMsg.GetOrder()))
	poolInfo := k.GetPoolInfo(ctx, market)
	require.EqualValues(t, 1000*100, poolInfo.MoneyOrderBookReserve.Int64(), "money amount in order book should be 0")

	sellOrderMsg := msgCreateOrder
	sellOrderMsg.Side = types.ASK
	sellOrderMsg.Sender = to
	require.NoError(t, k.AddLimitOrder(ctx, sellOrderMsg.GetOrder()))

	// 2. check pool for full deal orders
	poolInfo = k.GetPoolInfo(ctx, market)
	require.EqualValues(t, 0, poolInfo.StockOrderBookReserve.Int64(), "stock amount in order book should be 0")
	require.EqualValues(t, 0, poolInfo.MoneyOrderBookReserve.Int64(), "money amount in order book should be 0")

	// 3. check account balance
	afterMoneyBalanceFrom := app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	afterStockBalanceFrom := app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	afterMoneyBalanceTo := app.AccountKeeper.GetAccount(ctx, to).GetCoins().AmountOf(moneySymbol)
	afterStockBalanceTo := app.AccountKeeper.GetAccount(ctx, to).GetCoins().AmountOf(stockSymbol)
	fmt.Println(beforeStockBalanceFrom, beforeMoneyBalanceFrom, afterStockBalanceFrom, afterMoneyBalanceFrom)
	fmt.Println(beforeStockBalanceTo, beforeMoneyBalanceTo, afterStockBalanceTo, afterMoneyBalanceTo)
	makeFee := buyOrderMsg.Quantity * param.MakerFeeRateRate / types.DefaultFeePrecision                        // will charge stock
	takerFee := sellOrderMsg.Quantity * sellOrderMsg.Price * param.TakerFeeRateRate / types.DefaultFeePrecision // will charge money
	fmt.Println(makeFee, takerFee)
	require.EqualValues(t, 100*1000, beforeMoneyBalanceFrom.Sub(afterMoneyBalanceFrom).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 1000-makeFee, afterStockBalanceFrom.Sub(beforeStockBalanceFrom).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 100*1000-takerFee, afterMoneyBalanceTo.Sub(beforeMoneyBalanceTo).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 1000, beforeStockBalanceTo.Sub(afterStockBalanceTo).Int64(), "balance in account isn't correct")

	// ------- half deal

	// get users balance
	beforeMoneyBalanceFrom = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	beforeStockBalanceFrom = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	beforeMoneyBalanceTo = app.AccountKeeper.GetAccount(ctx, to).GetCoins().AmountOf(moneySymbol)
	beforeStockBalanceTo = app.AccountKeeper.GetAccount(ctx, to).GetCoins().AmountOf(stockSymbol)

	// 3. add buy, sell orders
	buyOrderMsg.Identify = 2
	require.NoError(t, k.AddLimitOrder(ctx, buyOrderMsg.GetOrder()))
	sellOrderMsg.Identify = 2
	sellOrderMsg.Quantity = 500
	require.NoError(t, k.AddLimitOrder(ctx, sellOrderMsg.GetOrder()))

	// 4. check pool info for half deal
	poolInfo = k.GetPoolInfo(ctx, market)
	require.EqualValues(t, 0, poolInfo.StockOrderBookReserve.Int64(), "stock amount in order book shoule be 0")
	require.EqualValues(t, 500*100, poolInfo.MoneyOrderBookReserve.Int64(), "money amount in order book shoule be 0")

	// 3. check account balance
	afterMoneyBalanceFrom = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	afterStockBalanceFrom = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	afterMoneyBalanceTo = app.AccountKeeper.GetAccount(ctx, to).GetCoins().AmountOf(moneySymbol)
	afterStockBalanceTo = app.AccountKeeper.GetAccount(ctx, to).GetCoins().AmountOf(stockSymbol)
	fmt.Println(beforeStockBalanceFrom, beforeMoneyBalanceFrom, afterStockBalanceFrom, afterMoneyBalanceFrom)
	fmt.Println(beforeStockBalanceTo, beforeMoneyBalanceTo, afterStockBalanceTo, afterMoneyBalanceTo)
	makeFee = buyOrderMsg.Quantity / 2 * param.MakerFeeRateRate / types.DefaultFeePrecision                    // will charge stock
	takerFee = sellOrderMsg.Quantity * sellOrderMsg.Price * param.TakerFeeRateRate / types.DefaultFeePrecision // will charge money
	fmt.Println(makeFee, takerFee)
	require.EqualValues(t, 100*1000, beforeMoneyBalanceFrom.Sub(afterMoneyBalanceFrom).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 500, beforeStockBalanceTo.Sub(afterStockBalanceTo).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 500-makeFee, afterStockBalanceFrom.Sub(beforeStockBalanceFrom).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 100*500-takerFee, afterMoneyBalanceTo.Sub(beforeMoneyBalanceTo).Int64(), "balance in account isn't correct")
}

func TestPairKeeper_AllocateFeeToValidatorAndPool(t *testing.T) {
	app := testapp.NewTestApp()
	ctx := app.NewCtx()
	app.AutoSwapKeeper.SetParams(ctx, types.DefaultParams())
	keeper := app.AutoSwapKeeper.IPairKeeper.(*keepers.PairKeeper)

	//setup accounts
	app.AccountKeeper.SetAccount(ctx, supply.NewEmptyModuleAccount(types.PoolModuleAcc))
	app.AccountKeeper.SetAccount(ctx, supply.NewEmptyModuleAccount(auth.FeeCollectorName))
	baseAcc := auth.NewBaseAccount(sender, sdk.NewCoins(sdk.NewCoin("money", sdk.NewInt(100)),
		sdk.NewCoin("stock", sdk.NewInt(100))), nil, 0, 0)
	app.AccountKeeper.SetAccount(ctx, baseAcc)

	_, err := keeper.AllocateFeeToValidatorAndPool(ctx, "money", sdk.NewInt(5), sender)
	require.Nil(t, err)

	_, err = keeper.AllocateFeeToValidatorAndPool(ctx, "stock", sdk.NewInt(10), sender)
	require.Nil(t, err)

	poolCoins := app.AccountKeeper.GetAccount(ctx, supply.NewModuleAddress(types.PoolModuleAcc)).GetCoins()
	feeCollectorCoins := app.AccountKeeper.GetAccount(ctx, supply.NewModuleAddress(auth.FeeCollectorName)).GetCoins()
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("money", sdk.NewInt(3)), sdk.NewCoin("stock", sdk.NewInt(6))), poolCoins)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("money", sdk.NewInt(2)), sdk.NewCoin("stock", sdk.NewInt(4))), feeCollectorCoins)
}

func TestPairKeeper_DealOrdersWitPool(t *testing.T) {
	var (
		app           = prepareTestApp(t)
		ctx           = app.ctx
		k             = app.AutoSwapKeeper
		reserveAmount = int64(100000000)
		param         = types.DefaultParams()
	)

	// 1. init pool with amounts
	require.NoError(t, k.SendCoinsFromAccountToModule(ctx, from, types.PoolModuleAcc, newCoins(stockSymbol, sdk.NewInt(reserveAmount))))
	require.NoError(t, k.SendCoinsFromAccountToModule(ctx, from, types.PoolModuleAcc, newCoins(moneySymbol, sdk.NewInt(reserveAmount))))
	fmt.Println(k.IPairKeeper.(*keepers.PairKeeper).GetAccount(ctx, supply.NewModuleAddress(types.PoolModuleAcc)).GetCoins().String())
	market := fmt.Sprintf("%s/%s", stockSymbol, moneySymbol)
	k.SetPoolInfo(ctx, market, &keepers.PoolInfo{
		Symbol:          market,
		MoneyAmmReserve: sdk.NewInt(reserveAmount),
		StockAmmReserve: sdk.NewInt(reserveAmount),
	})

	// 2. add order to deal with pool
	msgCreateOrder := types.MsgCreateOrder{
		TradingPair: market,
		Price:       100,
		Quantity:    1000,
		Side:        types.BID,
		Identify:    1,
		Sender:      from,
	}
	buyOrderMsg := msgCreateOrder

	beforePoolInfo := k.GetPoolInfo(ctx, market)
	currDealMoney := keepers.IntoPoolAmountTillPrice(sdk.NewDec(buyOrderMsg.Price), true, beforePoolInfo)
	if currDealMoney.GT(buyOrderMsg.GetOrder().ActualAmount()) {
		currDealMoney = buyOrderMsg.GetOrder().ActualAmount()
	}
	outAmount := keepers.GetAmountOutInPool(currDealMoney, beforePoolInfo, buyOrderMsg.Side == types.BID)
	totalFee := outAmount.Mul(sdk.NewInt(param.DealWithPoolFeeRate)).Quo(sdk.NewInt(types.DefaultFeePrecision))
	feeToPool := totalFee.Sub(totalFee.MulRaw(param.FeeToValidator).QuoRaw(param.FeeToValidator + param.FeeToPool))
	require.NoError(t, k.AddLimitOrder(ctx, buyOrderMsg.GetOrder()))
	fmt.Println(totalFee.String(), feeToPool.String(), outAmount.String())

	afterPoolInfo := k.GetPoolInfo(ctx, market)
	require.EqualValues(t, 0, afterPoolInfo.StockOrderBookReserve.Int64())
	require.EqualValues(t, 0, afterPoolInfo.MoneyOrderBookReserve.Int64())
	fmt.Println(beforePoolInfo.StockAmmReserve.String(), afterPoolInfo.StockAmmReserve.String())
	require.EqualValues(t, outAmount.Int64(), beforePoolInfo.StockAmmReserve.Sub(afterPoolInfo.StockAmmReserve).Int64())
	require.EqualValues(t, 100*1000, afterPoolInfo.MoneyAmmReserve.Sub(beforePoolInfo.MoneyAmmReserve).Int64())
}

func TestPairKeeper_DealOrdersWitPoolAndOrderBook(t *testing.T) {
	var (
		app           = prepareTestApp(t)
		ctx           = app.ctx
		k             = app.AutoSwapKeeper
		reserveAmount = int64(10000)
	)

	// 1. add orders
	market := fmt.Sprintf("%s/%s", stockSymbol, moneySymbol)
	msgCreateOrder := types.MsgCreateOrder{
		TradingPair: market,
		Price:       1,
		Quantity:    10000,
		Side:        types.BID,
		Identify:    1,
		Sender:      from,
	}
	buyOrderIDs := make([]string, 0)
	for i := 0; i < 10; i++ {
		tmpOrder := *msgCreateOrder.GetOrder()
		tmpOrder.Identify = byte(i)
		buyOrderIDs = append(buyOrderIDs, tmpOrder.GetOrderID())
		require.NoError(t, k.AddLimitOrder(ctx, &tmpOrder))
	}

	// 2. init pool with amounts
	require.NoError(t, k.SendCoinsFromAccountToModule(ctx, from, types.PoolModuleAcc, newCoins(stockSymbol, sdk.NewInt(reserveAmount))))
	require.NoError(t, k.SendCoinsFromAccountToModule(ctx, from, types.PoolModuleAcc, newCoins(moneySymbol, sdk.NewInt(reserveAmount))))
	fmt.Println(k.IPairKeeper.(*keepers.PairKeeper).GetAccount(ctx, supply.NewModuleAddress(types.PoolModuleAcc)).GetCoins().String())
	k.SetPoolInfo(ctx, market, &keepers.PoolInfo{
		Symbol:          market,
		MoneyAmmReserve: sdk.NewInt(reserveAmount),
		StockAmmReserve: sdk.NewInt(reserveAmount),
	})

	// 3. deal orders
	sellOrderMsg := msgCreateOrder
	sellOrderMsg.Sender = to
	sellOrderMsg.Identify = 1
	require.NoError(t, k.AddLimitOrder(ctx, sellOrderMsg.GetOrder()))
	fmt.Println(k.GetPoolInfo(ctx, market))
}

func TestPairKeeper_IntoPoolAmountTillPrice(t *testing.T) {
	require.EqualValues(t, 0,
		keepers.IntoPoolAmountTillPrice(sdk.NewDec(90), true,
			&keepers.PoolInfo{MoneyAmmReserve: sdk.NewInt(1000_000), StockAmmReserve: sdk.NewInt(10_000)}).Int64())
	require.EqualValues(t, 0,
		keepers.IntoPoolAmountTillPrice(sdk.NewDec(110), false,
			&keepers.PoolInfo{MoneyAmmReserve: sdk.NewInt(1000_000), StockAmmReserve: sdk.NewInt(10_000)}).Int64())
	require.EqualValues(t, 0,
		keepers.IntoPoolAmountTillPrice(sdk.NewDec(100), true,
			&keepers.PoolInfo{MoneyAmmReserve: sdk.NewInt(1000_000), StockAmmReserve: sdk.NewInt(10_000)}).Int64())
	require.EqualValues(t, 0,
		keepers.IntoPoolAmountTillPrice(sdk.NewDec(100), false,
			&keepers.PoolInfo{MoneyAmmReserve: sdk.NewInt(1000_000), StockAmmReserve: sdk.NewInt(10_000)}).Int64())
	//require.EqualValues(t, 48797,
	//	keepers.IntoPoolAmountTillPrice(sdk.NewDec(110), true,
	//		&keepers.PoolInfo{MoneyAmmReserve: sdk.NewInt(1000_000), StockAmmReserve: sdk.NewInt(10_000)}).Int64())
	require.EqualValues(t, 540,
		keepers.IntoPoolAmountTillPrice(sdk.NewDec(90), false,
			&keepers.PoolInfo{MoneyAmmReserve: sdk.NewInt(1000_000), StockAmmReserve: sdk.NewInt(10_000)}).Int64())
}
