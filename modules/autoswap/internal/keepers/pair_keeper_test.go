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

	// 1. first set poolInfo
	market := fmt.Sprintf("%s/%s", stockSymbol, moneySymbol)
	k.SetPoolInfo(ctx, market, &keepers.PoolInfo{
		Symbol:                market,
		StockAmmReserve:       sdk.ZeroInt(),
		MoneyAmmReserve:       sdk.ZeroInt(),
		StockOrderBookReserve: sdk.ZeroInt(),
		MoneyOrderBookReserve: sdk.ZeroInt(),
	})

	// 2. add orders
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
	for i := 0; i < 10; i++ {
		tmpOrder := *msgCreateOrder.GetOrder()
		tmpOrder.Identify = byte(i)
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

	// -----------

	// 6. add set order
	beforeMoneyBalance = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	beforeStockBalance = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	msgCreateOrder.Side = types.ASK
	msgCreateOrder.Price = 200
	for i := 0; i < 10; i++ {
		tmpOrder := *msgCreateOrder.GetOrder()
		tmpOrder.Identify = byte(i)
		require.NoError(t, k.AddLimitOrder(ctx, &tmpOrder))
	}

	// 7. check orderIndex
	require.EqualValues(t, 21, k.OrderIndexInOneBlock())

	// 4. check poolInfo
	poolInfo = k.GetPoolInfo(ctx, market)
	require.EqualValues(t, poolInfo.StockAmmReserve.Int64(), 0, "stock pool reserve should be 0")
	require.EqualValues(t, poolInfo.MoneyAmmReserve.Int64(), 0, "stock pool reserve should be 0")
	require.EqualValues(t, int64(10*100*1000), poolInfo.MoneyOrderBookReserve.Int64(), "orderBook money amount isn't correct")
	require.EqualValues(t, int64(10*1000), poolInfo.StockOrderBookReserve.Int64(), "orderBook stock amount isn't correct")

	// 5. check account balance
	afterMoneyBalance = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(moneySymbol)
	afterStockBalance = app.AccountKeeper.GetAccount(ctx, from).GetCoins().AmountOf(stockSymbol)
	require.EqualValues(t, 0, beforeMoneyBalance.Sub(afterMoneyBalance).Int64(), "balance in account isn't correct")
	require.EqualValues(t, 10*1000, beforeStockBalance.Sub(afterStockBalance).Int64(), "balance in account isn't correct")
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

	err := keeper.AllocateFeeToValidatorAndPool(ctx, "money", sdk.NewInt(5), sender)
	require.Nil(t, err)

	err = keeper.AllocateFeeToValidatorAndPool(ctx, "stock", sdk.NewInt(10), sender)
	require.Nil(t, err)

	poolCoins := app.AccountKeeper.GetAccount(ctx, supply.NewModuleAddress(types.PoolModuleAcc)).GetCoins()
	feeCollectorCoins := app.AccountKeeper.GetAccount(ctx, supply.NewModuleAddress(auth.FeeCollectorName)).GetCoins()
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("money", sdk.NewInt(3)), sdk.NewCoin("stock", sdk.NewInt(6))), poolCoins)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("money", sdk.NewInt(2)), sdk.NewCoin("stock", sdk.NewInt(4))), feeCollectorCoins)
}
