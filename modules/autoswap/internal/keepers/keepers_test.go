package keepers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMint(t *testing.T) {
	addr := sdk.AccAddress("add123")
	th := newTestHelper(t)

	pair := th.createPair(addr, "foo", "bar")
	pair.mint(10000, 1000000, addr)
	require.Equal(t, sdk.NewInt(100000), pair.getLiquidity(addr))
	reserves := pair.getReserves()
	require.Equal(t, 10000, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney)
}

// contract("pair", async accounts => {
func TestPair(t *testing.T) {
	boss := sdk.AccAddress("boss")
	maker := sdk.AccAddress("maker")
	taker := sdk.AccAddress("taker")
	shareReceiver := sdk.AccAddress("shareReceiver")
	th := newTestHelper(t)

	// it("initialize pair with btc/usd", async () => {
	btc := th.issueToken("btc0", 100000000000000, boss)
	usd := th.issueToken("usd0", 100000000000000, boss)
	pair := th.createPair(maker, btc.sym, usd.sym)

	// it("mint", async () => {
	pair.mint(10000, 1000000, shareReceiver)
	reserves := pair.getReserves()
	require.Equal(t, 10000, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney)

	// it("insert sell order with 0 deal", async () => {
	btc.transfer(maker, 10000, boss)
	usd.transfer(taker, 1000000, boss)
	pair.addLimitOrder(false, maker, 100, makePrice32(10000000, 18), 1, merge3(0, 0, 0))
	pair.addLimitOrder(false, maker, 100, makePrice32(10300000, 18), 2, merge3(0, 0, 0))
	pair.addLimitOrder(false, maker, 100, makePrice32(10500000, 18), 3, merge3(2, 0, 0))
	pair.addLimitOrder(false, maker, 100, makePrice32(10700000, 18), 4, merge3(3, 0, 0))
	pair.addLimitOrder(false, maker, 100, makePrice32(10900000, 18), 5, merge3(4, 0, 0))
	pair.addLimitOrder(false, maker, 1, makePrice32(10200000, 18), 6, merge3(1, 0, 0))
	pair.addLimitOrder(false, maker, 1, makePrice32(10400000, 18), 7, merge3(2, 0, 0))
	pair.addLimitOrder(false, maker, 1, makePrice32(10600000, 18), 8, merge3(3, 0, 0))
	pair.addLimitOrder(false, maker, 1, makePrice32(10800000, 18), 9, merge3(4, 0, 0))
	require.Equal(t, 9496, btc.balanceOf(maker))
	require.Equal(t, 0, usd.balanceOf(maker))
	reserves = pair.getReserves()
	require.Equal(t, 10000, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney) // TODO
	require.Equal(t, 1, reserves.firstSellID)        // TODO
	booked := pair.getBooked()
	require.Equal(t, 504, booked.bookedStock) // TODO
	require.Equal(t, 0, booked.bookedMoney)
	require.Equal(t, 0, booked.firstBuyID) // TODO
	//th.getOrderList(pair, true)                  // TODO
	//th.getOrderList(pair, false)                 // TODO

	// it("insert buy order with only 1 incomplete deal with orderbook", async () => {
	usd.transfer(taker, 5000, boss)
	pair.addLimitOrder(true, taker, 50, makePrice32(10000000, 18), 11, merge3(0, 0, 0))
	require.Equal(t, 9995000, usd.balanceOf(taker))
	require.Equal(t, 49, btc.balanceOf(taker))
	reserves = pair.getReserves()
	require.Equal(t, 10001, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney)
	require.Equal(t, 1, reserves.firstSellID)
	booked = pair.getBooked()
	require.Equal(t, 454, booked.bookedStock)
	require.Equal(t, 0, booked.bookedMoney)
	require.Equal(t, 0, booked.firstBuyID)

	// it("insert buy order with only 1 complete deal with orderbook", async () => {
	usd.transfer(taker, 5100, boss)
	pair.addLimitOrder(true, taker, 51, makePrice32(10000000, 18), 12, merge3(0, 0, 0))
	reserves = pair.getReserves()
	require.Equal(t, 10001, reserves.reserveStock)
	require.Equal(t, 1000100, reserves.reserveMoney)
	require.Equal(t, 6, reserves.firstSellID)
	booked = pair.getBooked()
	require.Equal(t, 404, booked.bookedStock)
	require.Equal(t, 0, booked.bookedMoney)
	require.Equal(t, 0, booked.firstBuyID)
	require.Equal(t, 9989900, usd.balanceOf(taker))
	require.Equal(t, 99, btc.balanceOf(taker))

	// it("insert buy order with 7 complete deal with orderbook and 4 swap", async () => {
	usd.transfer(taker, 99000, boss)
	pair.addLimitOrder(true, taker, 900, makePrice32(11000000, 18), 12, merge3(0, 0, 0))
	reserves = pair.getReserves()
	require.Equal(t, 9538, reserves.reserveStock)
	require.Equal(t, 1049020, reserves.reserveMoney)
	require.Equal(t, 0, reserves.firstSellID)
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedStock)
	require.Equal(t, 7260, booked.bookedMoney)
	require.Equal(t, 12, booked.firstBuyID)
	require.Equal(t, 9890900, usd.balanceOf(taker))
	require.Equal(t, 966, btc.balanceOf(taker))

	// it("remove sell order", async () => {
	pair.removeOrder(true, 12, merge3(0, 0, 0), taker)
	reserves = pair.getReserves()
	require.Equal(t, 9538, reserves.reserveStock)
	require.Equal(t, 1049020, reserves.reserveMoney)
	require.Equal(t, 0, reserves.firstSellID)
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedStock)
	require.Equal(t, 0, booked.bookedMoney)
	require.Equal(t, 0, booked.firstBuyID)
	require.Equal(t, 9898160, usd.balanceOf(taker))
	require.Equal(t, 966, btc.balanceOf(taker))
}

// contract("insert & delete order", async (accounts) => {
func TestInsertAndDeleteOrder(t *testing.T) {
	// TODO
}

// contract("swap on low liquidity", async (accounts) => {
func TestSwapOnLowLiquidity(t *testing.T) {
	// TODO
}

// contract("big deal on low liquidity", async (accounts) => {
func TestBigDealOnLowLiquidity(t *testing.T) {
	// TODO
}

// contract("deal with pool", async (accounts) => {
func TestDealWithPool(t *testing.T) {
	// TODO
}

// contract("deal after donate and sync", async (accounts) => {
func TestDealAfterDonateAndSync(t *testing.T) {
	// TODO
}

// contract("pair with weth token", async (accounts) => {
func TestPairWithWethToken(t *testing.T) {
	// TODO
}

// contract("pair with eth token", async (accounts) => {
func TestPairWithEthToken(t *testing.T) {
	// TODO
}

// contract("OneSwapPair/addMarketOrder", async (accounts) => {
func TestAddMarketOrder(t *testing.T) {
	// TODO
}

// contract("OneSwapPair/addMarketOrder/emptyAMM", async (accounts) => {
func TestEmptyAMM(t *testing.T) {
	// TODO
}

func TestAddMarketOrderEat(t *testing.T) {
	// TODO
}
