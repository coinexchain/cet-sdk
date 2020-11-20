package keepers_test

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"

	"github.com/coinexchain/cet-sdk/modules/asset"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/assert"

	testapp "github.com/coinexchain/cet-sdk/testapp"
	"github.com/stretchr/testify/require"

	"github.com/coinexchain/cet-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	testNum              = 5000
	stockSymbol          = "tokenone"
	moneySymbol          = "tokentwo"
	moduleAcc            = "autoswap-pool"
	maxTokenAmount int64 = 1000000
	from                 = testutil.ToAccAddress("bob")
	to                   = testutil.ToAccAddress("alice")
	tokenAmount          = sdk.NewInt(1e18).Mul(sdk.NewInt(1e18))
)

func getRandom(max int64) sdk.Int {
	return sdk.NewInt(rand.Int63n(max))
}

type App struct {
	*testapp.TestApp
	ctx sdk.Context
}

func prepareTestApp(t *testing.T) *App {
	app := testapp.NewTestApp()
	ctx := app.NewCtx()
	issueTokenAndTransfer(ctx, app, t)

	market := fmt.Sprintf("%s/%s", stockSymbol, moneySymbol)
	app.AutoSwapKeeper.SetPoolInfo(ctx, market, &keepers.PoolInfo{
		Symbol:                market,
		StockAmmReserve:       sdk.ZeroInt(),
		MoneyAmmReserve:       sdk.ZeroInt(),
		StockOrderBookReserve: sdk.ZeroInt(),
		MoneyOrderBookReserve: sdk.ZeroInt(),
	})
	app.AutoSwapKeeper.SetParams(ctx, types.DefaultParams())
	return &App{app, ctx}
}

func issueTokenAndTransfer(ctx sdk.Context, app *testapp.TestApp, t *testing.T) {
	app.SupplyKeeper.SetSupply(ctx, supply.Supply{Total: sdk.Coins{}})
	app.AssetKeeper.SetParams(ctx, asset.DefaultParams())

	err := app.AssetKeeper.IssueToken(ctx, stockSymbol, stockSymbol, tokenAmount, from,
		false, false, false, false, "", "", "123")
	assert.Nil(t, err)
	err = app.AssetKeeper.IssueToken(ctx, moneySymbol, moneySymbol, tokenAmount, from,
		false, false, false, false, "", "", "456")
	assert.Nil(t, err)

	fromAcc := app.AccountKeeper.NewAccountWithAddress(ctx, from)
	toAcc := app.AccountKeeper.NewAccountWithAddress(ctx, to)
	require.NoError(t, fromAcc.SetCoins(newCoins(stockSymbol, tokenAmount.Quo(sdk.NewInt(2))).Add(
		newCoins(moneySymbol, tokenAmount.Quo(sdk.NewInt(2))))), "set coins to account failed ")
	require.NoError(t, toAcc.SetCoins(newCoins(stockSymbol, tokenAmount.Quo(sdk.NewInt(2))).Add(
		newCoins(moneySymbol, tokenAmount.Quo(sdk.NewInt(2))))), "set coins to account failed ")
	app.AccountKeeper.SetAccount(ctx, fromAcc)
	app.AccountKeeper.SetAccount(ctx, toAcc)
}

func newCoins(token string, amount sdk.Int) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(token, amount))
}

func TestLiquidity(t *testing.T) {
	var (
		market = fmt.Sprintf("%s/%s", stockSymbol, moneySymbol)
	)
	app := prepareTestApp(t)
	app.AutoSwapKeeper.SetPoolInfo(app.ctx, market, &keepers.PoolInfo{Symbol: market, PricePrecision: 18})
	mintLiquidityTest(t, app, market)
	burnLiquidityTest(t, app, market)
	addLimitOrderTest(t, app, market)
}

func mintLiquidityTest(t *testing.T, app *App, market string) {
	stockAmount := getRandom(maxTokenAmount).Mul(sdk.NewInt(1e18))
	moneyAmount := getRandom(maxTokenAmount).Mul(sdk.NewInt(1e18))
	if !app.AccountKeeper.GetAccount(app.ctx, from).GetCoins().IsAllGT(newCoins(stockSymbol, stockAmount).Add(newCoins(moneySymbol, moneyAmount))) {
		fmt.Println("The random amount is larger than the user's balance")
		return
	}
	// transfer amount of token to moduleAccount
	require.Nil(t, app.AutoSwapKeeper.SendCoinsFromAccountToModule(app.ctx, from, moduleAcc,
		newCoins(stockSymbol, stockAmount)), "transfer stock coins from account to module failed")
	require.Nil(t, app.AutoSwapKeeper.SendCoinsFromAccountToModule(app.ctx, from, moduleAcc,
		newCoins(moneySymbol, moneyAmount)), "transfer money coins from account to module failed")
	// mint liquidity and check added liquidity
	expectedLiquidity := sdk.NewIntFromBigInt(big.NewInt(0).Sqrt(stockAmount.Mul(moneyAmount).BigInt()))
	_, err := app.AutoSwapKeeper.Mint(app.ctx, market, stockAmount, moneyAmount, to)
	require.Nil(t, err, "init liquidity mint failed")
	info := app.AutoSwapKeeper.GetPoolInfo(app.ctx, market)
	require.EqualValues(t, info.TotalSupply, expectedLiquidity, "liquidity is not equal")

	for i := 0; i < testNum; i++ {
		// get random amount to addLiquidity
		info = app.AutoSwapKeeper.GetPoolInfo(app.ctx, market)
		stockAmount, moneyAmount := info.GetLiquidityAmountIn(getRandom(
			maxTokenAmount).Mul(sdk.NewInt(1e18)), getRandom(maxTokenAmount).Mul(sdk.NewInt(1e18)))
		if !app.AccountKeeper.GetAccount(app.ctx, from).GetCoins().IsAllGT(newCoins(stockSymbol, stockAmount).Add(newCoins(moneySymbol, moneyAmount))) {
			fmt.Println("The random amount is larger than the user's balance")
			continue
		}

		// transfer amount of token to moduleAccount
		require.Nil(t, app.AutoSwapKeeper.SendCoinsFromAccountToModule(app.ctx, from, moduleAcc,
			newCoins(stockSymbol, stockAmount)), "transfer stock coins from account to module failed")
		require.Nil(t, app.AutoSwapKeeper.SendCoinsFromAccountToModule(app.ctx, from, moduleAcc,
			newCoins(moneySymbol, moneyAmount)), "transfer money coins from account to module failed")

		// mint liquidity
		beforeLiquidity := info.TotalSupply
		expectedLiquidity = getExpectedLiquidity(stockAmount, moneyAmount, info)
		if i%2 == 0 {
			_, err = app.AutoSwapKeeper.Mint(app.ctx, market, stockAmount, moneyAmount, to)
			require.Nil(t, err, "subsequent liquidity mint failed")
		} else {
			_, err := app.AutoSwapKeeper.Mint(app.ctx, market,
				stockAmount, moneyAmount, from)
			require.Nil(t, err, "subsequent liquidity mint failed")
		}

		// check added liquidity
		info = app.AutoSwapKeeper.GetPoolInfo(app.ctx, market)
		require.EqualValues(t, info.TotalSupply.Sub(beforeLiquidity), expectedLiquidity, "subsequent liquidity is not equal")
	}
}

func getExpectedLiquidity(stockAmount, moneyAmount sdk.Int, info *keepers.PoolInfo) sdk.Int {
	liquidity := stockAmount.Mul(info.TotalSupply).Quo(info.StockAmmReserve)
	another := moneyAmount.Mul(info.TotalSupply).Quo(info.MoneyAmmReserve)
	if liquidity.LT(another) {
		return liquidity
	}
	return another
}

func burnLiquidityTest(t *testing.T, app *App, market string) {
	// get random liquidity to burn
	ctx := app.ctx
	burnLiquidityAmount := getRandom(maxTokenAmount).Mul(sdk.NewInt(1e9))
	if app.AutoSwapKeeper.GetLiquidity(ctx, market, to).LT(burnLiquidityAmount) {
		fmt.Println("The random liquidity amount is larger than the user's balance")
		return
	}
	info := app.AutoSwapKeeper.GetPoolInfo(app.ctx, market)
	expectedStockOut, expectedMoneyOut := info.GetTokensAmountOut(burnLiquidityAmount)

	// burn liquidity
	stockOut, moneyOut, err := app.AutoSwapKeeper.Burn(app.ctx, market, from, burnLiquidityAmount)
	require.Nil(t, err, "init liquidity burn failed")
	// check outToken is correct
	require.EqualValues(t, stockOut, expectedStockOut, "get stock amount is not equal in burn liquidity")
	require.EqualValues(t, moneyOut, expectedMoneyOut, "get money amount is not equal in burn liquidity")
	// todo. check token balance in from address

	// check result
	for i := 0; i < testNum; i++ {
		burnLiqudityAmount := getRandom(maxTokenAmount).Mul(sdk.NewInt(1e9))
		info = app.AutoSwapKeeper.GetPoolInfo(app.ctx, market)
		expectedStockOut, expectedMoneyOut = info.GetTokensAmountOut(burnLiqudityAmount)
		// todo. transfer token to moduleAccount
		stockOut, moneyOut, err = app.AutoSwapKeeper.Burn(app.ctx, market, from, burnLiqudityAmount)
		require.Nil(t, err, "subsequent liquidity burn failed")
		// check outToken is correct
		require.EqualValues(t, stockOut.String(), expectedStockOut.String(), "get stock amount is not equal in burn liquidity")
		require.EqualValues(t, moneyOut.String(), expectedMoneyOut.String(), "get money amount is not equal in burn liquidity")
		// check liquidity balance in to address
	}
}

func addLimitOrderTest(t *testing.T, app *App, market string) {
	stockBalance := app.AccountKeeper.GetAccount(app.ctx, from).GetCoins().AmountOf(stockSymbol)
	moneyBalance := app.AccountKeeper.GetAccount(app.ctx, to).GetCoins().AmountOf(moneySymbol)
	require.NoError(t, app.BankKeeper.SendCoins(app.ctx, from, to, newCoins(stockSymbol, stockBalance.Quo(sdk.NewInt(2)))), "transfer stock token failed")
	require.NoError(t, app.BankKeeper.SendCoins(app.ctx, from, to, newCoins(moneySymbol, moneyBalance.Quo(sdk.NewInt(2)))), "transfer money token failed")
	fmt.Println("from coins: ", app.AccountKeeper.GetAccount(app.ctx, from).GetCoins())
	for i := 0; i < testNum; i++ {
		quantity := getRandom(100000000)
		order := &types.Order{
			TradingPair: market,
			Sender:      from,
			IsBuy:       true,
			Price:       getRandomPrice(100000, 18),
			Identify:    byte(i),
			Sequence:    int64(i),
			Quantity:    quantity.Int64(),
			Height:      int64(i),
			LeftStock:   quantity.Int64(),
		}
		token := order.Stock()
		if i%2 == 0 {
			order.Sender = to
			order.IsBuy = false
			token = order.Money()
		}
		if !app.AccountKeeper.GetAccount(app.ctx, order.Sender).GetCoins().AmountOf(token).GT(order.ActualAmount()) {
			continue
		}
		fmt.Println(order.ActualAmount(), "; orderID: ", order.GetOrderID())
		err := app.AutoSwapKeeper.AddLimitOrder(app.ctx, order)
		if err != nil {
			require.EqualValues(t, types.CodeUnKnownError, err.Code())
		}

		fmt.Println("Add order ok ...")
	}
}

func getRandomPrice(maxPrice, maxPrecision int64) sdk.Dec {
	price := getRandom(maxPrice)
	pricePrecision := getRandom(maxPrecision)
	if pricePrecision.IsZero() {
		pricePrecision = sdk.NewInt(1)
	}
	return sdk.NewDecFromInt(price).Quo(sdk.NewDec(int64(math.Pow10(int(pricePrecision.Int64())))))
}
