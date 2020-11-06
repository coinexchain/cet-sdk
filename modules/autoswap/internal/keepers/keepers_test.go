package keepers_test

import (
	"fmt"
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
	usd.transfer(taker, 10000000, boss)
	fee := pair.th.app.AutoSwapKeeper.GetDealWithPoolFee(pair.th.ctx)
	fmt.Println(fee)

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
	require.Equal(t, 1000000, reserves.reserveMoney)
	require.Equal(t, 1, reserves.firstSellID)
	booked := pair.getBooked()
	require.Equal(t, 504, booked.bookedStock)
	require.Equal(t, 0, booked.bookedMoney)
	require.Equal(t, -1, booked.firstBuyID)
	//th.getOrderList(pair, true)                  // TODO
	//th.getOrderList(pair, false)                 // TODO

	// it("insert buy order with only 1 incomplete deal with orderbook", async () => {
	//usd.transfer(taker, 5000, boss)
	pair.addLimitOrder(true, taker, 50, makePrice32(10000000, 18), 11, merge3(0, 0, 0))
	require.Equal(t, 9995000, usd.balanceOf(taker)) // TODO
	require.Equal(t, 49, btc.balanceOf(taker))      // TODO
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

	// it("insert sell order with duplicated id", async () => {
	btc.transfer(maker, 100, boss)
	pair.addLimitOrder(false, maker, 100, makePrice32(11000000, 18), 1, merge3(0, 0, 0))
	btc.transfer(maker, 50, boss)
	pair.addLimitOrder(false, maker, 50, makePrice32(12000000, 18), 1, merge3(0, 0, 0))

	// it("insert sell order with invalid prevkey", async () => {
	btc.transfer(maker, 100, boss)
	pair.addLimitOrder(false, maker, 100, makePrice32(10500000, 18), 1, merge3(1, 2, 3))

	// it("insert buy order with invalid price", async () => {
	// "OneSwap: INVALID_PRICE"
	pair.addLimitOrder(true, taker, 100, makePrice32(105000, 18), 1, merge3(1, 2, 3))
	// "OneSwap: INVALID_PRICE"
	pair.addLimitOrder(true, taker, 100, makePrice32(105000000, 18), 1, merge3(1, 2, 3))

	// it("insert buy order with unenough usd", async () => {
	usd.transfer(taker, 500, boss)
	// "OneSwap: DEPOSIT_NOT_ENOUGH"
	pair.addLimitOrder(true, taker, 50, makePrice32(10000000, 18), 1, merge3(0, 0, 0))

	// it("remove buy order with non existed id", async () => {
	// "OneSwap: NO_SUCH_ORDER"
	pair.removeOrder(true, 1, merge3(0, 0, 0), taker)

	// it("remove sell order with invalid prevKey", async () => {
	// "OneSwap: REACH_END"
	pair.removeOrder(false, 1, merge3(2, 3, 3), maker)

	// it("only order sender can remove order", async () => {
	// "OneSwap: NOT_OWNER"
	pair.removeOrder(false, 3, merge3(0, 0, 0), boss)

	// it("remove sell order successfully", async () => {
	// remove first sell id emits sync event at first
	pair.removeOrder(false, 3, merge3(0, 0, 0), maker)
	pair.removeOrder(false, 2, merge3(1, 0, 0), maker)
	pair.removeOrder(false, 1, merge3(0, 0, 0), maker)
}

// contract("swap on low liquidity", async (accounts) => {
func TestSwapOnLowLiquidity(t *testing.T) {
	boss := sdk.AccAddress("boss")
	maker := sdk.AccAddress("maker")
	//taker := sdk.AccAddress("taker")
	shareReceiver := sdk.AccAddress("shareReceiver")
	//lp := sdk.AccAddress("lp")
	th := newTestHelper(t)

	// it("initialize pair with btc/usd", async () => {
	btc := th.issueToken("btc0", 100000000000000, boss)
	usd := th.issueToken("usd0", 100000000000000, boss)
	pair := th.createPair(maker, btc.sym, usd.sym)

	// it("mint only 1000 shares", async () => {
	btc.transfer(shareReceiver, 10000, boss)
	usd.transfer(shareReceiver, 1000000, boss)
	pair.mint(10000, 1000000, shareReceiver)
	//balance := pair.balanceOf(shareReceiver)
	// TODO

	// it("swap with pool", async () => {
	btc.transfer(maker, 90000000000000, boss)
	pair.addMarketOrder(false, maker, 90000000000000)
	// TODO
}

// contract("big deal on low liquidity", async (accounts) => {
func TestBigDealOnLowLiquidity(t *testing.T) {
	boss := sdk.AccAddress("boss")
	maker := sdk.AccAddress("maker")
	taker := sdk.AccAddress("taker")
	shareReceiver := sdk.AccAddress("shareReceiver")
	//lp := sdk.AccAddress("lp")
	th := newTestHelper(t)

	// it("initialize pair with btc/usd", async () => {
	btc := th.issueToken("btc0", 100000000000000, boss)
	usd := th.issueToken("usd0", 100000000000000, boss)
	pair := th.createPair(maker, btc.sym, usd.sym)

	// it("mint only 1000 shares", async () => {
	btc.transfer(shareReceiver, 10000, boss)
	usd.transfer(shareReceiver, 1000000, boss)
	pair.mint(10000, 1000000, shareReceiver)
	//balance := pair.balanceOf(shareReceiver)
	// TODO

	// it("insert sell order at pool current price", async () => {
	btc.transfer(maker, 90000000000000, boss)
	pair.addLimitOrder(false, maker, 10, makePrice32(10000000, 18), 1, merge3(0, 0, 0))
	require.Equal(t, 1, usd.balanceOf(maker))
	booked := pair.getBooked()
	require.Equal(t, 0, booked.bookedMoney)
	require.Equal(t, 10, booked.bookedStock)

	// it("insert three small buy order at lower price", async () => {
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedMoney)
	pair.addLimitOrder(true, maker, 10, makePrice32(20000000, 16), 1, merge3(0, 0, 0))
	booked = pair.getBooked()
	require.Equal(t, 20, booked.bookedMoney)
	pair.addLimitOrder(true, maker, 10, makePrice32(30000000, 16), 1, merge3(0, 0, 0))
	booked = pair.getBooked()
	require.Equal(t, 50, booked.bookedMoney)
	pair.addLimitOrder(true, maker, 10, makePrice32(40000000, 16), 1, merge3(0, 0, 0))
	booked = pair.getBooked()
	require.Equal(t, 90, booked.bookedMoney)

	// it("insert big sell order not deal", async () => {
	pair.addLimitOrder(false, maker, 1000000000, makePrice32(10100000, 18), 2, merge3(0, 0, 0))
	require.Equal(t, 0, usd.balanceOf(maker))
	booked = pair.getBooked()
	require.Equal(t, 90, booked.bookedMoney)
	require.Equal(t, 1000000010, booked.bookedStock)
	require.Equal(t, 1, pair.getReserves().firstSellID)

	// it("insert big order deal with biggest sell order ", async () => {
	usd.transfer(taker, 100000_0000_0000, boss)
	pair.addLimitOrder(true, taker, 10_0000_0000, makePrice32(10100000, 18), 1, merge3(0, 0, 0))
	balance := btc.balanceOf(taker)
	require.Equal(t, 9_9700_0000, balance)

	// it("insert big buy order to hao yang mao", async () => {
	pair.addMarketOrder(true, taker, 10_0000)
	balanceAfter := btc.balanceOf(taker)
	require.Equal(t, 2492375, balanceAfter-balance)

	// it("insert sell order", async () => {
	balance = usd.balanceOf(taker)
	booked = pair.getBooked()
	pair.addMarketOrder(false, taker, 100)
	balanceAfter = usd.balanceOf(taker)
	require.Equal(t, 105, balanceAfter-balance)
}

// contract("deal with pool", async (accounts) => {
func TestDealWithPool(t *testing.T) {
	boss := sdk.AccAddress("boss")
	maker := sdk.AccAddress("maker")
	//taker := sdk.AccAddress("taker")
	shareReceiver := sdk.AccAddress("shareReceiver")
	//lp := sdk.AccAddress("lp")
	th := newTestHelper(t)

	// it("initialize pair with btc/usd", async () => {
	btc := th.issueToken("btc0", 100000000000000, boss)
	usd := th.issueToken("usd0", 100000000000000, boss)
	pair := th.createPair(maker, btc.sym, usd.sym)

	// it("mint only 1000 shares", async () => {
	btc.transfer(shareReceiver, 10000, boss)
	usd.transfer(shareReceiver, 1000000, boss)
	pair.mint(10000, 1000000, shareReceiver)
	//balance := pair.balanceOf(shareReceiver)
	// TODO

	//  it("insert buy order which can not be dealt", async ()=>{
	pair.addLimitOrder(true, maker, 100, makePrice32(1000_0000, 18), 1, merge3(0, 0, 0))
	// TODO

	// it("insert sell order which can deal totally", async ()=>{
	pair.addLimitOrder(false, maker, 10, makePrice32(9000_0000, 17), 1, merge3(0, 0, 0))
	// TODO

	// it("insert sell order which eats all buy order", async ()=>{
	pair.addLimitOrder(false, maker, 90, makePrice32(1000_0000, 18), 1, merge3(0, 0, 0))
	// TODO

	// it("insert buy order which can not be dealt", async ()=>{
	pair.addLimitOrder(true, maker, 10, makePrice32(1010_0000, 18), 1, merge3(0, 0, 0))
	reserves := pair.getReserves()
	require.Equal(t, 100, reserves.reserveStock)
	require.Equal(t, 10131, reserves.reserveMoney)
	// TODO
}

// contract("deal after donate and sync", async (accounts) => {
func TestDealAfterDonateAndSync(t *testing.T) {
	// TODO

	// it("mint only 1000 shares", async () => {

	// it("deal with pool", async () => {
}

// contract("pair with weth token", async (accounts) => {
func TestPairWithWethToken(t *testing.T) {
	// TODO

	// it("mint only 1000 shares", async () => {

	// it("insert buy order which can not be dealt", async ()=>{

	// it("insert sell order which can deal totally", async ()=>{

	// it("insert sell order which eats up buy order", async ()=>{

	// it("addMarketOrder to buy eth", async ()=>{

	// it("addMarketOrder to sell eth", async ()=>{
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

// contract("OneSwapPair/addMarketOrder/eat", async (accounts) => {
func TestAddMarketOrderEat(t *testing.T) {
	// TODO
}
