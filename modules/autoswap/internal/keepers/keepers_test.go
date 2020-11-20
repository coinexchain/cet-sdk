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

	// it("insert buy order with only 1 incomplete deal with orderbook", async () => {
	//usd.transfer(taker, 5000, boss)
	require.Equal(t, 10000000, usd.balanceOf(taker))
	require.Equal(t, 0, btc.balanceOf(taker))
	pair.addLimitOrder(true, taker, 50, 100, 11)
	require.Equal(t, 9995000, usd.balanceOf(taker))
	require.Equal(t, 49, btc.balanceOf(taker))
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
	fmt.Println("usd balance: ", usd.balanceOf(taker), "; btc balance: ", btc.balanceOf(taker))
	pair.addLimitOrder(true, taker, 51, 100, 12)
	reserves = pair.getReserves()
	require.Equal(t, 10002, reserves.reserveStock)   // oneswap 10001, because the more fee to pool due to charging mechanism.
	require.Equal(t, 1000000, reserves.reserveMoney) // oneswap 1000100, because no deal with pool.
	booked = pair.getBooked()
	require.Equal(t, 404, booked.bookedStock)
	require.Equal(t, 100, booked.bookedMoney)       // oneswap 0, because left 100 money no deal with orderBook and pool.
	require.Equal(t, 9995000, usd.balanceOf(taker)) // oneswap 9989900, because only deal with orderBook amount = 50, left 1.
	require.Equal(t, 98, btc.balanceOf(taker))      // because sub 1 fee in dealOrderBook

	// it("insert buy order with 7 complete deal with orderbook and 4 swap", async () => {
	usd.transfer(taker, 99000, boss)
	fmt.Println("before usd : ", usd.balanceOf(taker))
	pair.addLimitOrder(true, taker, 900, 110, 13)
	reserves = pair.getReserves()
	require.Equal(t, 9546, reserves.reserveStock)    // oneswap 9538, because more precision, dealInfos[-99, -48, -48, -47, -46, -45, -45, -44, -44], +10 fee
	require.Equal(t, 1048913, reserves.reserveMoney) // oneswap 1049020, dealInfos[+10051, +4939, +4915, +4892, +4868, +4846, +4823, +4801, +4778]
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedStock)
	require.Equal(t, 7367, booked.bookedMoney)      // oneswap 7260. before + order.ActualAmount - amountInToPool
	require.Equal(t, 9995000, usd.balanceOf(taker)) // oneswap 9890900, because freeze user money, but freeze amount is not as useful token.
	require.Equal(t, 958, btc.balanceOf(taker))     // oneswap 966, because charging mechanism: 12 fee stock

	// it("remove sell order", async () => {
	order := types.Order{
		Sender:   taker,
		Sequence: 0,
		Identify: 13,
	}
	pair.removeOrder(order.GetOrderID(), taker)
	reserves = pair.getReserves()
	require.Equal(t, 9546, reserves.reserveStock)    // oneswap 9538,
	require.Equal(t, 1048913, reserves.reserveMoney) // oneswap 1049020,
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedStock)
	require.Equal(t, 100, booked.bookedMoney)        // left order.Price = 100, amount = 100.
	require.Equal(t, 10002267, usd.balanceOf(taker)) // oneswap 9898160, before(9995000) + leftFreeze(7267) = 10002267
	require.Equal(t, 958, btc.balanceOf(taker))      // oneswap 966, same before.
	fmt.Println(th.app.AccountKeeper.GetAccount(th.ctx, supply.NewModuleAddress(types.PoolModuleAcc)).GetCoins())
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
	btc.transfer(supply.NewModuleAddress(types.PoolModuleAcc), 10000, boss)
	usd.transfer(supply.NewModuleAddress(types.PoolModuleAcc), 1000000, boss)
	reserves := pair.getReserves()
	require.Equal(t, 10000, reserves.reserveStock)
	require.Equal(t, 1000000, reserves.reserveMoney)

	// it("insert sell order with duplicated id", async () => {
	btc.transfer(maker, 100, boss)
	pair.addLimitOrder(false, maker, 100, 110, 1)
	require.EqualValues(t, 1, len(th.app.AutoSwapKeeper.GetAllOrders(th.ctx, "btc0/usd0")))
	btc.transfer(maker, 50, boss)
	pair.addLimitOrder(false, maker, 50, 120, 2)
	require.EqualValues(t, 2, len(th.app.AutoSwapKeeper.GetAllOrders(th.ctx, "btc0/usd0")))
	// The test case is remove, because there is no prevKey in the current design.
	// so now only insert orders.
	// it("insert sell order with invalid prevkey", async () => {
	btc.transfer(maker, 100, boss)
	pair.addLimitOrder(false, maker, 100, 105, 3)
	require.EqualValues(t, 3, len(th.app.AutoSwapKeeper.GetAllOrders(th.ctx, "btc0/usd0")))

	// it("insert buy order with invalid price", async () => {
	// "OneSwap: INVALID_PRICE"
	//pair.addLimitOrder(true, taker, 100, makePrice32(105000, 18), 1) // TODO
	// "OneSwap: INVALID_PRICE"
	//pair.addLimitOrder(true, taker, 100, makePrice32(105000000, 18), 1) // TODO

	// it("insert buy order with unenough usd", async () => {
	usd.transfer(taker, 500, boss)
	// "OneSwap: DEPOSIT_NOT_ENOUGH"
	err := pair.addLimitOrderWithoutCheck(true, taker, 50, 100, 1)
	require.NotNil(t, err)
	require.EqualValues(t, sdk.CodeInsufficientCoins, err.Code())

	// it("remove buy order with non existed id", async () => {
	// "OneSwap: NO_SUCH_ORDER"
	order := types.Order{
		Sequence: int64(th.app.AccountKeeper.GetAccount(th.ctx, taker).GetSequence()),
		Identify: 1,
		Sender:   taker,
	}
	err = pair.removeOrderWithoutCheck(order.GetOrderID(), taker)
	require.NotNil(t, err)
	require.EqualValues(t, types.CodeInvalidOrderID, err.Code())

	//  The test case is remove, because there is no prevKey in the current design.
	// it("remove sell order with invalid prevKey", async () => {
	// "OneSwap: REACH_END"
	//order.Sender = maker
	//order.Sequence = int64(th.app.AccountKeeper.GetAccount(th.ctx, maker).GetSequence())
	//pair.removeOrder(order.GetOrderID(), maker)

	// it("only order sender can remove order", async () => {
	// "OneSwap: NOT_OWNER"
	order = types.Order{
		Sequence: 0,
		Identify: 3,
		Sender:   maker,
	}
	err = pair.removeOrderWithoutCheck(order.GetOrderID(), boss)
	require.NotNil(t, err)
	require.EqualValues(t, types.CodeInvalidOrderSender, err.Code())

	// it("remove sell order successfully", async () => {
	// remove first sell id emits sync event at first
	order = types.Order{
		Sequence: 0,
		Identify: 3,
		Sender:   maker,
	}
	order.Sequence = int64(th.app.AccountKeeper.GetAccount(th.ctx, maker).GetSequence())
	pair.removeOrder(order.GetOrderID(), maker)
	order = types.Order{
		Sequence: 0,
		Identify: 2,
		Sender:   maker,
	}
	pair.removeOrder(order.GetOrderID(), maker)
	order = types.Order{
		Sequence: 0,
		Identify: 1,
		Sender:   maker,
	}
	pair.removeOrder(order.GetOrderID(), maker)
}

// Delete the test case. because no market order.
// contract("swap on low liquidity", async (accounts) => {
//func TestSwapOnLowLiquidity(t *testing.T) {
//	boss := sdk.AccAddress("boss")
//	maker := sdk.AccAddress("maker")
//	//taker := sdk.AccAddress("taker")
//	shareReceiver := sdk.AccAddress("shareReceiver")
//	//lp := sdk.AccAddress("lp")
//	th := newTestHelper(t)
//
//	// it("initialize pair with btc/usd", async () => {
//	btc := th.issueToken("btc0", 100000000000000, boss)
//	usd := th.issueToken("usd0", 100000000000000, boss)
//	pair := th.createPair(maker, btc.sym, usd.sym, 0)
//
//	// it("mint only 1000 shares", async () => {
//	btc.transfer(shareReceiver, 10000, boss)
//	usd.transfer(shareReceiver, 1000000, boss)
//	pair.mint(10000, 1000000, shareReceiver)
//	//balance := pair.balanceOf(shareReceiver)
//	// TODO
//
//	// it("swap with pool", async () => {
//	btc.transfer(maker, 90000000000000, boss)
//	pair.addMarketOrder(false, maker, 90000000000000)
//	// TODO
//}

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
	btc.transfer(supply.NewModuleAddress(types.PoolModuleAcc), 10000, boss)
	usd.transfer(supply.NewModuleAddress(types.PoolModuleAcc), 1000000, boss)
	//balance := pair.balanceOf(shareReceiver)
	// TODO

	// it("insert sell order at pool current price", async () => {
	btc.transfer(maker, 90000000000000, boss)
	pair.addLimitOrder(false, maker, 10, 100, 1)
	require.Equal(t, 0, usd.balanceOf(maker))
	booked := pair.getBooked()
	require.Equal(t, 0, booked.bookedMoney)
	require.Equal(t, 10, booked.bookedStock)

	// it("insert three small buy order at lower price", async () => {
	booked = pair.getBooked()
	require.Equal(t, 0, booked.bookedMoney)
	usd.transfer(maker, 20, boss)
	pair.addLimitOrder(true, maker, 10, 2, 2)
	booked = pair.getBooked()
	require.Equal(t, 20, booked.bookedMoney)
	usd.transfer(maker, 30, boss)
	pair.addLimitOrder(true, maker, 10, 3, 3)
	booked = pair.getBooked()
	require.Equal(t, 50, booked.bookedMoney)
	usd.transfer(maker, 40, boss)
	pair.addLimitOrder(true, maker, 10, 4, 4)
	booked = pair.getBooked()
	require.Equal(t, 90, booked.bookedMoney)

	// it("insert big sell order not deal", async () => {
	//btc.transfer(pair, 1000000000, maker)
	pair.addLimitOrder(false, maker, 1000000000, 101, 5)
	require.Equal(t, 0, usd.balanceOf(maker))
	booked = pair.getBooked()
	require.Equal(t, 90, booked.bookedMoney)
	require.Equal(t, 1000000010, booked.bookedStock)

	// it("insert big order deal with biggest sell order ", async () => {
	usd.transfer(taker, 100000_0000_0000, boss)
	pair.addLimitOrder(true, taker, 10_0000_0000, 101, 6)
	balance := btc.balanceOf(taker)
<<<<<<< HEAD
	require.Equal(t, 997000057, balance) //  oneswap: 9_9699_9999
=======
	require.Equal(t, 996999998, balance) //  oneswap: 9_9699_9999
>>>>>>> c73774b4... Pass all test;

	// it("insert big buy order to hao yang mao", async () => {
	//pair.addMarketOrder(true, taker, 10_0000, 7)
	//balanceAfter := btc.balanceOf(taker)
	//require.Equal(t, 2492375, balanceAfter-balance)
	//
	//// it("insert sell order", async () => {
	//balance = usd.balanceOf(taker)
	//booked = pair.getBooked()
	//pair.addMarketOrder(false, taker, 100, 8)
	//balanceAfter = usd.balanceOf(taker)
	//require.Equal(t, 105, balanceAfter-balance)
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
	btc.transfer(supply.NewModuleAddress(types.PoolModuleAcc), 10000, boss)
	usd.transfer(supply.NewModuleAddress(types.PoolModuleAcc), 1000000, boss)
	//balance := pair.balanceOf(shareReceiver)
	// TODO

	//  it("insert buy order which can not be dealt", async ()=>{
	//pair.addLimitOrder(true, maker, 100, 100, 1)
	//// TODO
	//
	//// it("insert sell order which can deal totally", async ()=>{
	//pair.addLimitOrder(false, maker, 10, 90, 1)
	//// TODO
	//
	//// it("insert sell order which eats all buy order", async ()=>{
	//pair.addLimitOrder(false, maker, 90, 100, 1)
	//// TODO
	//
	//// it("insert buy order which can not be dealt", async ()=>{
	//pair.addLimitOrder(true, maker, 10, 101, 1)
	//reserves := pair.getReserves()
	//require.Equal(t, 100, reserves.reserveStock)
	//require.Equal(t, 10131, reserves.reserveMoney)
	// TODO
}
