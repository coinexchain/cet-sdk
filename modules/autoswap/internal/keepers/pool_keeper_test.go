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
		IsSwapOpen:            true,
		IsOrderBookOpen:       true,
		StockAmmReserve:       sdk.NewInt(100),
		MoneyAmmReserve:       sdk.NewInt(10000),
		StockOrderBookReserve: sdk.ZeroInt(),
		MoneyOrderBookReserve: sdk.ZeroInt(),
		TotalSupply:           sdk.ZeroInt(),
		KLast:                 sdk.ZeroInt(),
	}
	k := app.AutoSwapKeeper
	//step1: set pool info, stock/money = 100/10000
	k.SetPoolInfo(ctx, marketKey, true, true, &poolInfo)
	info := k.GetPoolInfo(ctx, marketKey, true, true)
	require.NotNil(t, info)
	require.Equal(t, info.Symbol, marketKey)
	require.Equal(t, info.StockAmmReserve.Int64(), int64(100))
	require.Equal(t, info.MoneyAmmReserve.Int64(), int64(10000))
	require.True(t, info.IsSwapOpen)
	require.True(t, info.IsOrderBookOpen)
	//step2: set pool info with swap close
	poolInfo.IsSwapOpen = false
	k.SetPoolInfo(ctx, marketKey, false, true, &poolInfo)
	info = k.GetPoolInfo(ctx, marketKey, false, true)
	require.NotNil(t, info)
	require.Equal(t, marketKey, info.Symbol)
	require.Equal(t, int64(100), info.StockAmmReserve.Int64())
	require.Equal(t, int64(10000), info.MoneyAmmReserve.Int64())
	require.False(t, info.IsSwapOpen)
	require.True(t, info.IsOrderBookOpen)
	//step3: get pool not exist
	info = k.GetPoolInfo(ctx, marketKey, true, false)
	require.Nil(t, info)
	var bear = sdk.AccAddress("bear")
	//step4: set liquidity
	var liquidity = sdk.NewInt(100)
	k.SetLiquidity(ctx, marketKey, true, true, bear, liquidity)
	l := k.GetLiquidity(ctx, marketKey, true, true, bear)
	require.Equal(t, liquidity.Int64(), l.Int64())
	//step5: get liquidity in pool not exist
	l = k.GetLiquidity(ctx, marketKey, false, true, bear)
	require.Equal(t, int64(0), l.Int64())
	//step6: clear liquidity
	k.ClearLiquidity(ctx, marketKey, true, true, bear)
	l = k.GetLiquidity(ctx, marketKey, true, true, bear)
	require.Equal(t, int64(0), l.Int64())
	//step7 mint
	//k.Mint(ctx, marketKey, true, true, sdk.NewInt(10), sdk.NewInt(100), bear)

}
