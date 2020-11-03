package keepers_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/coinexchain/cet-sdk/modules/asset"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/assert"

	testapp "github.com/coinexchain/cet-sdk/testapp"
	"github.com/stretchr/testify/require"

	"github.com/coinexchain/cet-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	testNum      = 19000
	tokenNameOne = "token_one"
	tokenNameTwo = "token_two"
	from         = testutil.ToAccAddress("bob")
	to           = testutil.ToAccAddress("alice")
	tokenAmount  = sdk.NewInt(1e18).Mul(sdk.NewInt(1e18))
)

func getRandom(max int64) sdk.Int {
	for i := 0; i < 10; i++ {
		fmt.Println(rand.Int63n(max))
	}
	return sdk.Int{}
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

	err := app.AssetKeeper.IssueToken(ctx, tokenNameOne, tokenNameOne, tokenAmount, from,
		false, false, false, false, "", "", "123")
	assert.Nil(t, err)
	err = app.AssetKeeper.IssueToken(ctx, tokenNameTwo, tokenNameTwo, tokenAmount, from,
		false, false, false, false, "", "", "456")
	assert.Nil(t, err)

	bobAcc := app.AccountKeeper.NewAccountWithAddress(ctx, from)
	//_ = bobAcc.SetCoins(newCoins(tokenNameOne, tokenAmount))
	//_ = bobAcc.SetCoins(newCoins(tokenNameTwo, tokenAmount))
	app.AccountKeeper.SetAccount(ctx, bobAcc)
}

func newCoins(token string, amount sdk.Int) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(token, amount))
}

func TestLiquidity(t *testing.T) {
	var (
		market          = fmt.Sprintf("%s/%s", tokenNameOne, tokenNameTwo)
		isOpenSwap      = true
		isOpenOrderBook = true
	)
	app := prepareTestApp(t)
	mintLiquidityTest(t, app, market, isOpenSwap, isOpenOrderBook)
}

func mintLiquidityTest(t *testing.T, app *App, market string, isOpenSwap, isOpenOrderBook bool) {
	var maxTokenAmount = sdk.ZeroInt()

	// todo. transfer token to moduleAccount
	// mint
	err := app.AutoSwapKeeper.Mint(app.ctx, market, isOpenSwap, isOpenOrderBook, sdk.ZeroInt(), sdk.ZeroInt(), to)
	require.Nil(t, err, "init liquidity mint failed")
	// check liquidity balance in to address

	// check result
	for i := 0; i < testNum; i++ {
		info := app.AutoSwapKeeper.GetPoolInfo(app.ctx, market, isOpenSwap, isOpenOrderBook)
		stockAmount, moneyAmount := info.GetLiquidityAmountIn(getRandom(maxTokenAmount.Int64()), getRandom(maxTokenAmount.Int64()))

		// todo. transfer token to moduleAccount
		err = app.AutoSwapKeeper.Mint(app.ctx, market, isOpenSwap, isOpenOrderBook, stockAmount, moneyAmount, to)
		require.Nil(t, err, "subsequent liquidity mint failed")
		// check liquidity balance in to address
	}
}

func burnLiquidityTest(t *testing.T, app *App, market string, isOpenSwap, isOpenOrderBook bool) {
	var maxLiquidity int64 = 0

	// todo. burn liqudity
	burnLiqudityAmount := getRandom(maxLiquidity)
	info := app.AutoSwapKeeper.GetPoolInfo(app.ctx, market, isOpenSwap, isOpenOrderBook)
	expectedStockOut, expectedMoneyOut := info.GetTokensAmountOut(burnLiqudityAmount)
	stockOut, moneyOut, err := app.AutoSwapKeeper.Burn(app.ctx, market, isOpenSwap, isOpenOrderBook, from, burnLiqudityAmount)
	require.Nil(t, err, "init liquidity burn failed")
	// check outToken is correct
	require.EqualValues(t, stockOut, expectedStockOut, "get stock amount is not equal in burn liquidity")
	require.EqualValues(t, moneyOut, expectedMoneyOut, "get money amount is not equal in burn liquidity")
	// todo. check token balance in from address

	// check result
	for i := 0; i < testNum; i++ {
		burnLiqudityAmount := getRandom(maxLiquidity)
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
