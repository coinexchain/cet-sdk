package keepers_test

import (
	"fmt"
	"testing"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMint(t *testing.T) {
	addr := sdk.AccAddress("add123")
	th := newTestHelper(t)

	pair := th.createPair(addr, "foo", "bar", 0)
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
	pair := th.createPair(maker, btc.sym, usd.sym, 0)

	// it("mint", async () => {
	pair.mint(10000, 1000000, shareReceiver)
	reserves := pair.getReserves()
	btc.transfer(supply.NewModuleAddress(types.PoolModuleAcc), 10000, boss)
	usd.transfer(supply.NewModuleAddress(types.PoolModuleAcc), 1000000, boss)
	require.Equal(t, 10000, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney)

	// it("insert sell order with 0 deal", async () => {
	btc.transfer(maker, 10000, boss)
	usd.transfer(taker, 10000000, boss)
	fee := pair.th.app.AutoSwapKeeper.GetDealWithPoolFee(pair.th.ctx)
	fmt.Println("takerFeeRate: ", pair.th.app.AutoSwapKeeper.GetTakerFee(pair.th.ctx), "; poolFeeRate: ", fee)
	fmt.Println(fee)

	require.Equal(t, 10000, btc.balanceOf(maker))
	pair.addLimitOrder(false, maker, 100, 100, 1)
	pair.addLimitOrder(false, maker, 100, 103, 2)
	pair.addLimitOrder(false, maker, 100, 105, 3)
	pair.addLimitOrder(false, maker, 100, 107, 4)
	pair.addLimitOrder(false, maker, 100, 109, 5)
	pair.addLimitOrder(false, maker, 1, 102, 6)
	pair.addLimitOrder(false, maker, 1, 104, 7)
	pair.addLimitOrder(false, maker, 1, 106, 8)
	pair.addLimitOrder(false, maker, 1, 108, 9)
	require.Equal(t, 9496, btc.balanceOf(maker)) // 10000 - 504
	require.Equal(t, 0, usd.balanceOf(maker))
	reserves = pair.getReserves()
	require.Equal(t, 10000, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney)
	booked := pair.getBooked()
	require.Equal(t, 504, booked.bookedStock)
	require.Equal(t, 0, booked.bookedMoney)
	//th.getOrderList(pair, true)                  // TODO
	//th.getOrderList(pair, false)                 // TODO

	// it("insert buy order with only 1 incomplete deal with orderbook", async () => {
	//usd.transfer(taker, 5000, boss)
	require.Equal(t, 10000000, usd.balanceOf(taker))
	require.Equal(t, 0, btc.balanceOf(taker))
	pair.addLimitOrder(true, taker, 50, 100, 11)
	require.Equal(t, 9995000, usd.balanceOf(taker)) // TODO
	require.Equal(t, 49, btc.balanceOf(taker))      // TODO
	reserves = pair.getReserves()
	require.Equal(t, 10001, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney)
	booked = pair.getBooked()
	require.Equal(t, 454, booked.bookedStock)
	require.Equal(t, 0, booked.bookedMoney)

	// it("insert buy order with only 1 complete deal with orderbook", async () => {
	reserves = pair.getReserves()
	fmt.Println("reserve stock: ", reserves.reserveStock, "; reserves money: ", reserves.reserveMoney)
	usd.transfer(taker, 5100, boss)
	pair.addLimitOrder(true, taker, 51, 100, 12)
	reserves = pair.getReserves()
	//require.Equal(t, 10001, reserves.reserveStock)
	//require.Equal(t, 1000100, reserves.reserveMoney)
	booked = pair.getBooked()
	require.Equal(t, 404, booked.bookedStock)
	//require.Equal(t, 0, booked.bookedMoney) // todo. will check
	//require.Equal(t, 9989900, usd.balanceOf(taker))
	//require.Equal(t, 99, btc.balanceOf(taker))

	// it("insert buy order with 7 complete deal with orderbook and 4 swap", async () => {
	usd.transfer(taker, 99000, boss)
	pair.addLimitOrder(true, taker, 900, 110, 12)
	reserves = pair.getReserves()
	require.Equal(t, 9538, reserves.reserveStock)
	require.Equal(t, 1049020, reserves.reserveMoney)
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedStock)
	require.Equal(t, 7260, booked.bookedMoney)
	require.Equal(t, 9890900, usd.balanceOf(taker))
	require.Equal(t, 966, btc.balanceOf(taker))

	// it("remove sell order", async () => {
	pair.removeOrder(true, 12, taker)
	reserves = pair.getReserves()
	require.Equal(t, 9538, reserves.reserveStock)
	require.Equal(t, 1049020, reserves.reserveMoney)
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedStock)
	require.Equal(t, 0, booked.bookedMoney)
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
	pair := th.createPair(maker, btc.sym, usd.sym, 0)

	// it("mint", async () => {
	pair.mint(10000, 1000000, shareReceiver)
	reserves := pair.getReserves()
	require.Equal(t, 10000, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney)

	// it("insert sell order with duplicated id", async () => {
	btc.transfer(maker, 100, boss)
	require.NoError(t, pair.addLimitOrderWithoutCheck(false, maker, 100, 110, 1))
	btc.transfer(maker, 50, boss)
	err := pair.addLimitOrderWithoutCheck(false, maker, 50, 120, 1)
	require.NotNil(t, err)
	require.EqualValues(t, types.CodeOrderAlreadyExist, err.Code())

	// The test case is remove, because there is no prevKey in the current design.
	// it("insert sell order with invalid prevkey", async () => {
	//btc.transfer(maker, 100, boss)
	//pair.addLimitOrder(false, maker, 100, 105, 1)

	// it("insert buy order with invalid price", async () => {
	// "OneSwap: INVALID_PRICE"
	//pair.addLimitOrder(true, taker, 100, makePrice32(105000, 18), 1) // TODO
	// "OneSwap: INVALID_PRICE"
	//pair.addLimitOrder(true, taker, 100, makePrice32(105000000, 18), 1) // TODO

	// it("insert buy order with unenough usd", async () => {
	usd.transfer(taker, 500, boss)
	// "OneSwap: DEPOSIT_NOT_ENOUGH"
	pair.addLimitOrder(true, taker, 50, 100, 1)

	// it("remove buy order with non existed id", async () => {
	// "OneSwap: NO_SUCH_ORDER"
	pair.removeOrder(true, 1, taker)

	// it("remove sell order with invalid prevKey", async () => {
	// "OneSwap: REACH_END"
	pair.removeOrder(false, 1, maker)

	// it("only order sender can remove order", async () => {
	// "OneSwap: NOT_OWNER"
	pair.removeOrder(false, 3, boss)

	// it("remove sell order successfully", async () => {
	// remove first sell id emits sync event at first
	pair.removeOrder(false, 3, maker)
	pair.removeOrder(false, 2, maker)
	pair.removeOrder(false, 1, maker)
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
	pair := th.createPair(maker, btc.sym, usd.sym, 0)

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
	pair := th.createPair(maker, btc.sym, usd.sym, 0)

	// it("mint only 1000 shares", async () => {
	btc.transfer(shareReceiver, 10000, boss)
	usd.transfer(shareReceiver, 1000000, boss)
	pair.mint(10000, 1000000, shareReceiver)
	//balance := pair.balanceOf(shareReceiver)
	// TODO

	// it("insert sell order at pool current price", async () => {
	btc.transfer(maker, 90000000000000, boss)
	pair.addLimitOrder(false, maker, 10, 100, 1)
	require.Equal(t, 1, usd.balanceOf(maker))
	booked := pair.getBooked()
	require.Equal(t, 0, booked.bookedMoney)
	require.Equal(t, 10, booked.bookedStock)

	// it("insert three small buy order at lower price", async () => {
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedMoney)
	pair.addLimitOrder(true, maker, 10, 2, 1)
	booked = pair.getBooked()
	require.Equal(t, 20, booked.bookedMoney)
	pair.addLimitOrder(true, maker, 10, 3, 1)
	booked = pair.getBooked()
	require.Equal(t, 50, booked.bookedMoney)
	pair.addLimitOrder(true, maker, 10, 4, 1)
	booked = pair.getBooked()
	require.Equal(t, 90, booked.bookedMoney)

	// it("insert big sell order not deal", async () => {
	pair.addLimitOrder(false, maker, 1000000000, 101, 2)
	require.Equal(t, 0, usd.balanceOf(maker))
	booked = pair.getBooked()
	require.Equal(t, 90, booked.bookedMoney)
	require.Equal(t, 1000000010, booked.bookedStock)

	// it("insert big order deal with biggest sell order ", async () => {
	usd.transfer(taker, 100000_0000_0000, boss)
	pair.addLimitOrder(true, taker, 10_0000_0000, 101, 1)
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
	pair := th.createPair(maker, btc.sym, usd.sym, 0)

	// it("mint only 1000 shares", async () => {
	btc.transfer(shareReceiver, 10000, boss)
	usd.transfer(shareReceiver, 1000000, boss)
	pair.mint(10000, 1000000, shareReceiver)
	//balance := pair.balanceOf(shareReceiver)
	// TODO

	//  it("insert buy order which can not be dealt", async ()=>{
	pair.addLimitOrder(true, maker, 100, 100, 1)
	// TODO

	// it("insert sell order which can deal totally", async ()=>{
	pair.addLimitOrder(false, maker, 10, 90, 1)
	// TODO

	// it("insert sell order which eats all buy order", async ()=>{
	pair.addLimitOrder(false, maker, 90, 100, 1)
	// TODO

	// it("insert buy order which can not be dealt", async ()=>{
	pair.addLimitOrder(true, maker, 10, 101, 1)
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
