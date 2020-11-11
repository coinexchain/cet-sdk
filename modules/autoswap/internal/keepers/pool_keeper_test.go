package keepers_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/testapp"
)

func TestPoolKeeper_Pool(t *testing.T) {
	app := testapp.NewTestApp()
	ctx := sdk.NewContext(app.Cms, abci.Header{}, false, log.NewNopLogger())
	var marketKey = "stock/money"
	var poolInfo = keepers.PoolInfo{
		Symbol:                marketKey,
		StockAmmReserve:       sdk.NewInt(100),
		MoneyAmmReserve:       sdk.NewInt(10000),
		StockOrderBookReserve: sdk.ZeroInt(),
		MoneyOrderBookReserve: sdk.ZeroInt(),
		TotalSupply:           sdk.ZeroInt(),
	}
	k := app.AutoSwapKeeper
	//step1: set pool info, stock/money = 100/10000
	k.SetPoolInfo(ctx, marketKey, &poolInfo)
	info := k.GetPoolInfo(ctx, marketKey)
	require.NotNil(t, info)
	require.Equal(t, info.Symbol, marketKey)
	require.Equal(t, info.StockAmmReserve.Int64(), int64(100))
	require.Equal(t, info.MoneyAmmReserve.Int64(), int64(10000))
	//step2: set pool info with swap close
	k.SetPoolInfo(ctx, marketKey, &poolInfo)
	info = k.GetPoolInfo(ctx, marketKey)
	require.NotNil(t, info)
	require.Equal(t, marketKey, info.Symbol)
	require.Equal(t, int64(100), info.StockAmmReserve.Int64())
	require.Equal(t, int64(10000), info.MoneyAmmReserve.Int64())
	//step3: get pool exist
	info = k.GetPoolInfo(ctx, marketKey)
	require.NotNil(t, info)
	var bear = sdk.AccAddress("bear")
	//step4: set liquidity
	var liquidity = sdk.NewInt(100)
	k.SetLiquidity(ctx, marketKey, bear, liquidity)
	l := k.GetLiquidity(ctx, marketKey, bear)
	require.Equal(t, liquidity.Int64(), l.Int64())
	//step5: get liquidity in pool exist
	l = k.GetLiquidity(ctx, marketKey, bear)
	require.Equal(t, int64(100), l.Int64())
	//step6: clear liquidity
	k.ClearLiquidity(ctx, marketKey, bear)
	l = k.GetLiquidity(ctx, marketKey, bear)
	require.Equal(t, int64(0), l.Int64())

	/*TEST MINT*/
	//step1: reset pool info
	poolInfo.MoneyAmmReserve = sdk.ZeroInt()
	poolInfo.StockAmmReserve = sdk.ZeroInt()
	k.SetPoolInfo(ctx, marketKey, &poolInfo)
	_, err := k.Mint(ctx, marketKey, sdk.NewInt(100), sdk.NewInt(10000), bear)
	require.Nil(t, err)
	l = k.GetLiquidity(ctx, marketKey, bear)
	require.Equal(t, int64(1000), l.Int64())
	_, err = k.Mint(ctx, marketKey, sdk.NewInt(10), sdk.NewInt(100), bear)
	require.Nil(t, err)
	l = k.GetLiquidity(ctx, marketKey, bear)
	require.Equal(t, int64(1010), l.Int64())
	//step2: burn
	_, _, err = k.Burn(ctx, marketKey, bear, sdk.NewInt(2000))
	require.NotNil(t, err)
	info = k.GetPoolInfo(ctx, marketKey)
	require.Equal(t, int64(110), info.StockAmmReserve.Int64())
	require.Equal(t, int64(10100), info.MoneyAmmReserve.Int64())
	require.Equal(t, int64(1010), info.TotalSupply.Int64())
	stockOut, moneyOut, err := k.Burn(ctx, marketKey, bear, sdk.NewInt(1000))
	require.Nil(t, err)
	require.Equal(t, int64(108 /*1000/1010*110*/), stockOut.Int64())
	require.Equal(t, int64(10000), moneyOut.Int64())
	l = k.GetLiquidity(ctx, marketKey, bear)
	require.Equal(t, int64(10), l.Int64())
	stockOut, moneyOut, err = k.Burn(ctx, marketKey, bear, sdk.NewInt(10))
	require.Nil(t, err)
	require.Equal(t, int64(110-108), stockOut.Int64())
	require.Equal(t, int64(100), moneyOut.Int64())
	info = k.GetPoolInfo(ctx, marketKey)
	require.Equal(t, int64(0), info.StockAmmReserve.Int64())
	require.Equal(t, int64(0), info.MoneyAmmReserve.Int64())
	require.Equal(t, int64(0), info.TotalSupply.Int64())
	l = k.GetLiquidity(ctx, marketKey, bear)
	require.Equal(t, int64(0), l.Int64())

	/*TEST utils*/
	poolInfo.StockAmmReserve = sdk.NewInt(100)
	poolInfo.MoneyAmmReserve = sdk.NewInt(10000)
	poolInfo.TotalSupply = sdk.NewInt(1000)
	stockOut, moneyOut = poolInfo.GetTokensAmountOut(sdk.NewInt(100))
	require.Equal(t, int64(10), stockOut.Int64())
	require.Equal(t, int64(1000), moneyOut.Int64())
	stockR, moneyR := poolInfo.GetLiquidityAmountIn(sdk.NewInt(100), sdk.NewInt(100))
	require.Equal(t, int64(1), stockR.Int64())
	require.Equal(t, int64(100), moneyR.Int64())
}
