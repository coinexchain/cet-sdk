package keepers

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	testNum = 19000
)

func getRandom(max int64) sdk.Int {
	for i := 0; i < 10; i++ {
		fmt.Println(rand.Int63n(max))
	}
	return sdk.Int{}
}

type keepers struct {
	ctx            sdk.Context
	autoSwapKeeper Keeper
	from           sdk.AccAddress
	to             sdk.AccAddress
	// asset keeper
	// supply keeper
	// account keeper
}

func prepareKeepers(t *testing.T, market string) *keepers {
	return &keepers{}
}

func TestLiquidity(t *testing.T) {
	var (
		market          = "abc/cet"
		isOpenSwap      = true
		isOpenOrderBook = true
	)
	ks := prepareKeepers(t, market)

	mintLiquidityTest(t, ks, market, isOpenSwap, isOpenOrderBook)
}

func mintLiquidityTest(t *testing.T, ks *keepers, market string, isOpenSwap, isOpenOrderBook bool) {
	var maxTokenAmount = sdk.ZeroInt()

	// todo. transfer token to moduleAccount
	// mint
	err := ks.autoSwapKeeper.Mint(ks.ctx, market, isOpenSwap, isOpenOrderBook, sdk.ZeroInt(), sdk.ZeroInt(), ks.to)
	require.Nil(t, err, "init liquidity mint failed")
	// check liquidity balance in to address

	// check result
	for i := 0; i < testNum; i++ {
		info := ks.autoSwapKeeper.GetPoolInfo(ks.ctx, market, isOpenSwap, isOpenOrderBook)
		stockAmount, moneyAmount := info.GetLiquidityAmountIn(getRandom(maxTokenAmount.Int64()), getRandom(maxTokenAmount.Int64()))

		// todo. transfer token to moduleAccount
		err = ks.autoSwapKeeper.Mint(ks.ctx, market, isOpenSwap, isOpenOrderBook, stockAmount, moneyAmount, ks.to)
		require.Nil(t, err, "subsequent liquidity mint failed")
		// check liquidity balance in to address
	}
}

func burnLiquidityTest(t *testing.T, ks *keepers, market string, isOpenSwap, isOpenOrderBook bool) {
	var maxLiquidity int64 = 0

	// todo. burn liqudity
	burnLiqudityAmount := getRandom(maxLiquidity)
	info := ks.autoSwapKeeper.GetPoolInfo(ks.ctx, market, isOpenSwap, isOpenOrderBook)
	expectedStockOut, expectedMoneyOut := info.GetTokensAmountOut(burnLiqudityAmount)
	stockOut, moneyOut, err := ks.autoSwapKeeper.Burn(ks.ctx, market, isOpenSwap, isOpenOrderBook, ks.from, burnLiqudityAmount)
	require.Nil(t, err, "init liquidity burn failed")
	// check outToken is correct
	require.EqualValues(t, stockOut, expectedStockOut, "get stock amount is not equal in burn liquidity")
	require.EqualValues(t, moneyOut, expectedMoneyOut, "get money amount is not equal in burn liquidity")
	// todo. check token balance in from address

	// check result
	for i := 0; i < testNum; i++ {
		burnLiqudityAmount := getRandom(maxLiquidity)
		info = ks.autoSwapKeeper.GetPoolInfo(ks.ctx, market, isOpenSwap, isOpenOrderBook)
		expectedStockOut, expectedMoneyOut = info.GetTokensAmountOut(burnLiqudityAmount)
		// todo. transfer token to moduleAccount
		stockOut, moneyOut, err = ks.autoSwapKeeper.Burn(ks.ctx, market, isOpenSwap, isOpenOrderBook, ks.from, sdk.ZeroInt())
		require.Nil(t, err, "subsequent liquidity burn failed")
		// check outToken is correct
		require.EqualValues(t, stockOut, expectedStockOut, "get stock amount is not equal in burn liquidity")
		require.EqualValues(t, moneyOut, expectedMoneyOut, "get money amount is not equal in burn liquidity")
		// check liquidity balance in to address
	}
}

func TestHello(t *testing.T) {
	fmt.Println("hello")
}
