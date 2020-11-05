package keepers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMint(t *testing.T) {
	addr := sdk.AccAddress("add123")
	th := newTestHelper(t)

	th.createPair(addr, "foo", "bar")
	th.mint("foo/bar", 10000, 1000000, addr)
	require.Equal(t, sdk.NewInt(100000), th.getLiquidity("foo/bar", addr))
	pi := th.getPoolInfo("foo/bar")
	require.Equal(t, sdk.NewInt(10000), pi.StockAmmReserve)
	require.Equal(t, sdk.NewInt(1000000), pi.MoneyAmmReserve)
}

func TestPair(t *testing.T) {
	btc := "btc0"
	usd := "usd0"
	pair := "btc0/usd0"
	boss := sdk.AccAddress("boss")
	maker := sdk.AccAddress("maker")
	taker := sdk.AccAddress("taker")
	shareReceiver := sdk.AccAddress("shareReceiver")
	th := newTestHelper(t)

	th.issueToken(btc, 100000000000000, boss)
	th.issueToken(usd, 100000000000000, boss)
	th.createPair(maker, btc, usd)

	th.mint(pair, 10000, 1000000, shareReceiver)
	pi := th.getPoolInfo(pair)
	require.Equal(t, sdk.NewInt(10000), pi.StockAmmReserve)
	require.Equal(t, sdk.NewInt(1000000), pi.MoneyAmmReserve)

	// insert sell order with 0 deal
	th.transfer(btc, 10000, boss, maker)
	th.transfer(usd, 1000000, boss, taker)
	th.addLimitOrder(pair, false, maker, 100, makePrice32(10000000, 18), 1, merge3(0, 0, 0))
	th.addLimitOrder(pair, false, maker, 100, makePrice32(10300000, 18), 2, merge3(0, 0, 0))
	th.addLimitOrder(pair, false, maker, 100, makePrice32(10500000, 18), 3, merge3(2, 0, 0))
	th.addLimitOrder(pair, false, maker, 100, makePrice32(10700000, 18), 4, merge3(3, 0, 0))
	th.addLimitOrder(pair, false, maker, 100, makePrice32(10900000, 18), 5, merge3(4, 0, 0))
	th.addLimitOrder(pair, false, maker, 1, makePrice32(10200000, 18), 6, merge3(1, 0, 0))
	th.addLimitOrder(pair, false, maker, 1, makePrice32(10400000, 18), 7, merge3(2, 0, 0))
	th.addLimitOrder(pair, false, maker, 1, makePrice32(10600000, 18), 8, merge3(3, 0, 0))
	th.addLimitOrder(pair, false, maker, 1, makePrice32(10800000, 18), 9, merge3(4, 0, 0))
	require.Equal(t, sdk.NewInt(9496), th.balanceOf(btc, maker))
	require.Equal(t, sdk.NewInt(0), th.balanceOf(usd, maker))
	pi = th.getPoolInfo(pair)
	require.Equal(t, sdk.NewInt(10000), pi.StockAmmReserve)
	require.Equal(t, sdk.NewInt(1000000), pi.MoneyAmmReserve)   // TODO
	require.Equal(t, sdk.NewInt(504), pi.StockOrderBookReserve) // TODO
	require.Equal(t, sdk.NewInt(0), pi.StockOrderBookReserve)
	require.Equal(t, 1, th.getFirstSellID(pair)) // TODO
	require.Equal(t, 0, th.getFirstBuyID(pair))  // TODO
	th.getOrderList(pair, true)                  // TODO
	th.getOrderList(pair, false)                 // TODO

	// insert buy order with only 1 incomplete deal with orderbook
	th.transfer(usd, 5000, boss, taker)
	th.addLimitOrder(pair, true, taker, 50, makePrice32(10000000, 18), 11, merge3(0, 0, 0))
	require.Equal(t, sdk.NewInt(9995000), th.balanceOf(usd, taker))
	require.Equal(t, sdk.NewInt(49), th.balanceOf(btc, taker))
	pi = th.getPoolInfo(pair)
	require.Equal(t, sdk.NewInt(10001), pi.StockAmmReserve)
	require.Equal(t, sdk.NewInt(1000000), pi.MoneyAmmReserve)
	require.Equal(t, sdk.NewInt(454), pi.StockOrderBookReserve)
	require.Equal(t, sdk.NewInt(0), pi.MoneyOrderBookReserve)
	require.Equal(t, 1, th.getFirstSellID(pair))
	require.Equal(t, 0, th.getFirstBuyID(pair))

	// insert buy order with only 1 complete deal with orderbook
	th.transfer(usd, 5100, boss, taker)
	th.addLimitOrder(pair, true, taker, 51, makePrice32(10000000, 18), 12, merge3(0, 0, 0))
	pi = th.getPoolInfo(pair)
	require.Equal(t, sdk.NewInt(10001), pi.StockAmmReserve)
	require.Equal(t, sdk.NewInt(1000100), pi.MoneyAmmReserve)
	require.Equal(t, sdk.NewInt(404), pi.StockOrderBookReserve)
	require.Equal(t, sdk.NewInt(0), pi.MoneyOrderBookReserve)
	require.Equal(t, 6, th.getFirstSellID(pair))
	require.Equal(t, 0, th.getFirstBuyID(pair))
	require.Equal(t, sdk.NewInt(9989900), th.balanceOf(usd, taker))
	require.Equal(t, sdk.NewInt(99), th.balanceOf(btc, taker))

	// insert buy order with 7 complete deal with orderbook and 4 swap
	th.transfer(usd, 99000, boss, taker)
	th.addLimitOrder(pair, true, taker, 900, makePrice32(11000000, 18), 12, merge3(0, 0, 0))
	pi = th.getPoolInfo(pair)
	require.Equal(t, sdk.NewInt(9538), pi.StockAmmReserve)
	require.Equal(t, sdk.NewInt(1049020), pi.MoneyAmmReserve)
	require.Equal(t, sdk.NewInt(0), pi.StockOrderBookReserve)
	require.Equal(t, sdk.NewInt(7260), pi.MoneyOrderBookReserve)
	require.Equal(t, 0, th.getFirstSellID(pair))
	require.Equal(t, 12, th.getFirstBuyID(pair))
	require.Equal(t, sdk.NewInt(9890900), th.balanceOf(usd, taker))
	require.Equal(t, sdk.NewInt(966), th.balanceOf(btc, taker))

	// remove sell order
	th.removeOrder(pair, true, 12, merge3(0, 0, 0), taker)
	pi = th.getPoolInfo(pair)
	require.Equal(t, sdk.NewInt(9538), pi.StockAmmReserve)
	require.Equal(t, sdk.NewInt(1049020), pi.MoneyAmmReserve)
	require.Equal(t, sdk.NewInt(0), pi.StockOrderBookReserve)
	require.Equal(t, sdk.NewInt(0), pi.MoneyOrderBookReserve)
	require.Equal(t, 0, th.getFirstSellID(pair))
	require.Equal(t, 0, th.getFirstBuyID(pair))
	require.Equal(t, sdk.NewInt(9898160), th.balanceOf(usd, taker))
	require.Equal(t, sdk.NewInt(966), th.balanceOf(btc, taker))
}
