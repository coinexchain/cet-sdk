package keepers_test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"

	"github.com/coinexchain/cet-sdk/modules/asset"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/assert"

	testapp "github.com/coinexchain/cet-sdk/testapp"
	"github.com/stretchr/testify/require"

	"github.com/coinexchain/cet-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	testNum              = 19000
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
	issueAbbToBob(ctx, app, t)
	return &App{app, ctx}
}

func issueAbbToBob(ctx sdk.Context, app *testapp.TestApp, t *testing.T) {
	app.SupplyKeeper.SetSupply(ctx, supply.Supply{Total: sdk.Coins{}})
	app.AssetKeeper.SetParams(ctx, asset.DefaultParams())

	err := app.AssetKeeper.IssueToken(ctx, stockSymbol, stockSymbol, tokenAmount, from,
		false, false, false, false, "", "", "123")
	assert.Nil(t, err)
	err = app.AssetKeeper.IssueToken(ctx, moneySymbol, moneySymbol, tokenAmount, from,
		false, false, false, false, "", "", "456")
	assert.Nil(t, err)

	bobAcc := app.AccountKeeper.NewAccountWithAddress(ctx, from)
	require.Nil(t, bobAcc.SetCoins(newCoins(stockSymbol, tokenAmount).Add(
		newCoins(moneySymbol, tokenAmount))), "set coins to account failed ")
	app.AccountKeeper.SetAccount(ctx, bobAcc)
}

func newCoins(token string, amount sdk.Int) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(token, amount))
}

func TestLiquidity(t *testing.T) {
	var (
		market          = fmt.Sprintf("%s/%s", stock, money)
		isOpenSwap      = true
		isOpenOrderBook = true
	)
	app := prepareTestApp(t)
	app.AutoSwapKeeper.SetPoolInfo(app.ctx, market, isOpenSwap, isOpenOrderBook, &keepers.PoolInfo{Symbol: market})
	mintLiquidityTest(t, app, market, isOpenSwap, isOpenOrderBook)
	//burnLiquidityTest(t, app, market, isOpenSwap, isOpenOrderBook)
}

func mintLiquidityTest(t *testing.T, app *App, market string, isOpenSwap, isOpenOrderBook bool) {
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
	require.Nil(t, app.AutoSwapKeeper.Mint(app.ctx, market, isOpenSwap, isOpenOrderBook, stockAmount, moneyAmount, to), "init liquidity mint failed")
	info := app.AutoSwapKeeper.GetPoolInfo(app.ctx, market, isOpenSwap, isOpenOrderBook)
	require.EqualValues(t, info.TotalSupply, expectedLiquidity, "liquidity is not equal")

	for i := 0; i < testNum; i++ {
		// get random amount to addLiquidity
		info = app.AutoSwapKeeper.GetPoolInfo(app.ctx, market, isOpenSwap, isOpenOrderBook)
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
			require.Nil(t, app.AutoSwapKeeper.Mint(app.ctx, market, isOpenSwap, isOpenOrderBook,
				stockAmount, moneyAmount, to), "subsequent liquidity mint failed")
		} else {
			require.Nil(t, app.AutoSwapKeeper.Mint(app.ctx, market, isOpenSwap, isOpenOrderBook,
				stockAmount, moneyAmount, from), "subsequent liquidity mint failed")
		}

		// check added liquidity
		info = app.AutoSwapKeeper.GetPoolInfo(app.ctx, market, isOpenSwap, isOpenOrderBook)
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

func burnLiquidityTest(t *testing.T, app *App, market string, isOpenSwap, isOpenOrderBook bool) {
	// todo. burn liqudity
	burnLiquidityAmount := getRandom(maxTokenAmount).Mul(sdk.NewInt(1e9))
	info := app.AutoSwapKeeper.GetPoolInfo(app.ctx, market, isOpenSwap, isOpenOrderBook)
	expectedStockOut, expectedMoneyOut := info.GetTokensAmountOut(burnLiquidityAmount)
	stockOut, moneyOut, err := app.AutoSwapKeeper.Burn(app.ctx, market, isOpenSwap, isOpenOrderBook, from, burnLiquidityAmount)
	require.Nil(t, err, "init liquidity burn failed")
	// check outToken is correct
	require.EqualValues(t, stockOut, expectedStockOut, "get stock amount is not equal in burn liquidity")
	require.EqualValues(t, moneyOut, expectedMoneyOut, "get money amount is not equal in burn liquidity")
	// todo. check token balance in from address

	// check result
	for i := 0; i < testNum; i++ {
		burnLiqudityAmount := getRandom(maxTokenAmount).Mul(sdk.NewInt(1e9))
		info = app.AutoSwapKeeper.GetPoolInfo(app.ctx, market, isOpenSwap, isOpenOrderBook)
		expectedStockOut, expectedMoneyOut = info.GetTokensAmountOut(burnLiqudityAmount)
		// todo. transfer token to moduleAccount
		stockOut, moneyOut, err = app.AutoSwapKeeper.Burn(app.ctx, market, isOpenSwap, isOpenOrderBook, from, sdk.ZeroInt())
		require.Nil(t, err, "subsequent liquidity burn failed")
		// check outToken is correct
		require.EqualValues(t, stockOut, expectedStockOut, "get stock amount is not equal in burn liquidity")
		require.EqualValues(t, moneyOut, expectedMoneyOut, "get money amount is not equal in burn liquidity")
		// check liquidity balance in to address
	}
}
