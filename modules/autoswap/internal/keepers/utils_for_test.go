package keepers_test

import (
	"testing"

	dex "github.com/coinexchain/cet-sdk/types"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/coinexchain/cet-sdk/modules/asset"
	"github.com/coinexchain/cet-sdk/modules/autoswap"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cet-sdk/testapp"
)

/* TestHelper */

type TestHelper struct {
	t   *testing.T
	app *testapp.TestApp
	ctx sdk.Context
}

func newTestHelper(t *testing.T) TestHelper {
	app, ctx := newTestApp()
	return TestHelper{t, app, ctx}
}
func (h TestHelper) issueToken(sym string, totalSupply int64, owner sdk.AccAddress) Token {
	issueToken(h.t, h.app.AssetKeeper, h.ctx, sym, sdk.NewInt(totalSupply), owner)
	return Token{h, sym}
}
func (h TestHelper) createPair(owner sdk.AccAddress, stock, money string, pricePrecision byte) Pair {
	createPair(h.t, h.app.AutoSwapKeeper, h.ctx, owner, stock, money, pricePrecision)
	return Pair{h, stock + "/" + money}
}

/* Token */

type Token struct {
	th  TestHelper
	sym string
}

func (t Token) transfer(to sdk.AccAddress, amt int64, from sdk.AccAddress) {
	coins := sdk.NewCoins(sdk.NewCoin(t.sym, sdk.NewInt(amt)))
	err := t.th.app.BankxKeeper.SendCoins(t.th.ctx, from, to, coins)
	require.NoError(t.th.t, err)
}
func (t Token) balanceOf(addr sdk.AccAddress) int {
	return int(t.th.app.BankxKeeper.GetCoins(t.th.ctx, addr).AmountOf(t.sym).Int64())
}

/* Pair */

type Pair struct {
	th  TestHelper
	sym string
}
type PairReserves struct {
	reserveStock int
	reserveMoney int
}
type PairBooked struct {
	bookedStock int
	bookedMoney int
}

func (p Pair) mint(stockIn, moneyIn int64, to sdk.AccAddress) {
	mint(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, p.sym, sdk.NewInt(stockIn), sdk.NewInt(moneyIn), to)
}
func (p Pair) addLimitOrder(isBuy bool, sender sdk.AccAddress, amt, price int64, identify byte) {
	addLimitOrder(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, p.sym, isBuy, sender, amt, price, identify)
}
func (p Pair) addLimitOrderWithoutCheck(isBuy bool, sender sdk.AccAddress, amt, price int64, identify byte) sdk.Error {
	return addLimitOrderWithoutCheck(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, p.sym, isBuy, sender, amt, price, identify)
}
func (p Pair) addMarketOrder(isBuy bool, sender sdk.AccAddress, amt int64) {
	addMarketOrder(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, p.sym, isBuy, sender, amt)
}
func (p Pair) removeOrder(id string, sender sdk.AccAddress) {
	removeOrder(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, id, sender)
}
func (p Pair) removeOrderWithoutCheck(id string, sender sdk.AccAddress) sdk.Error {
	return removeOrderWithoutCheck(p.th.t, p.th.ctx, p.th.app.AutoSwapKeeper, id, sender)
}
func (p Pair) getLiquidity(addr sdk.AccAddress) sdk.Int {
	return p.th.app.AutoSwapKeeper.GetLiquidity(p.th.ctx, p.sym, addr)
}
func (p Pair) getPoolInfo() *autoswap.PoolInfo {
	return p.th.app.AutoSwapKeeper.GetPoolInfo(p.th.ctx, p.sym)
}

func (p Pair) getReserves() PairReserves {
	pi := p.getPoolInfo()
	return PairReserves{
		reserveStock: int(pi.StockAmmReserve.Int64()),
		reserveMoney: int(pi.MoneyAmmReserve.Int64()),
	}
}
func (p Pair) getBooked() PairBooked {
	pi := p.getPoolInfo()
	return PairBooked{
		bookedStock: int(pi.StockOrderBookReserve.Int64()),
		bookedMoney: int(pi.MoneyOrderBookReserve.Int64()),
	}
}
func (p Pair) getOrder(isBuy bool, orderID int64) *types.Order {
	return p.th.app.AutoSwapKeeper.GetOrder(p.th.ctx, "TODO")
}
func (p Pair) getOrderList(isBuy bool) []*types.Order {
	// TODO
	return nil
}
func (p Pair) balanceOf(addr sdk.AccAddress) int {
	lq := p.th.app.AutoSwapKeeper.GetLiquidity(p.th.ctx, p.sym, addr)
	return int(lq.Int64())
}

/* helper functions */

func newTestApp() (app *testapp.TestApp, ctx sdk.Context) {
	app = testapp.NewTestApp()
	ctx = sdk.NewContext(app.Cms, abci.Header{}, false, log.NewNopLogger())
	app.SupplyKeeper.SetSupply(ctx, supply.Supply{Total: sdk.Coins{}})
	app.AssetKeeper.SetParams(ctx, asset.DefaultParams())
	app.AutoSwapKeeper.SetParams(ctx, types.Params{
		TakerFeeRateRate:    30,
		MakerFeeRateRate:    0,
		DealWithPoolFeeRate: 30,
		FeeToPool:           1,
		FeeToValidator:      0,
	})
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccount(ctx, supply.NewEmptyModuleAccount(autoswap.PoolModuleAcc)))
	return
}

func issueToken(t *testing.T, ak asset.Keeper, ctx sdk.Context,
	sym string, totalSupply sdk.Int, owner sdk.AccAddress) {

	err := ak.IssueToken(ctx, sym, sym, totalSupply, owner, false, false, false, false, sym, sym, sym)
	require.NoError(t, err)

	err = ak.SendCoinsFromAssetModuleToAccount(ctx, owner, sdk.NewCoins(sdk.NewCoin(sym, totalSupply)))
	require.NoError(t, err)
}

func createPair(t *testing.T, ask *autoswap.Keeper, ctx sdk.Context,
	owner sdk.AccAddress, stock, money string, pricePrecision byte) {
	ask.CreatePair(ctx, owner, dex.GetSymbol(stock, money), pricePrecision)
}

func mint(t *testing.T, ask *autoswap.Keeper, ctx sdk.Context,
	pair string, stockIn, moneyIn sdk.Int, to sdk.AccAddress) {

	_, err := ask.Mint(ctx, pair, stockIn, moneyIn, to)
	require.NoError(t, err)
}

func addLimitOrderWithoutCheck(t *testing.T, ask *autoswap.Keeper, ctx sdk.Context,
	pair string, isBuy bool, sender sdk.AccAddress, amt, price int64, id byte) sdk.Error {
	msg := types.MsgCreateOrder{
		TradingPair: pair,
		Price:       price,
		Quantity:    amt,
		Identify:    id,
		Sender:      sender,
	}
	if isBuy {
		msg.Side = types.BID
	} else {
		msg.Side = types.ASK
	}

	return ask.AddLimitOrder(ctx, msg.GetOrder())
}

func addLimitOrder(t *testing.T, ask *autoswap.Keeper, ctx sdk.Context,
	pair string, isBuy bool, sender sdk.AccAddress, amt, price int64, id byte) {
	err := addLimitOrderWithoutCheck(t, ask, ctx, pair, isBuy, sender, amt, price, id)
	require.NoError(t, err)
}

// TODO
func addMarketOrder(t *testing.T, ask *autoswap.Keeper, ctx sdk.Context,
	pair string, isBuy bool, sender sdk.AccAddress, amt int64) {

	err := ask.AddLimitOrder(ctx, &types.Order{
		Sender:      sender,
		TradingPair: pair,
		IsBuy:       isBuy,
		Quantity:    amt,
		// TODO
	})
	require.NoError(t, err)
}

func removeOrderWithoutCheck(t *testing.T, ctx sdk.Context,
	ask *autoswap.Keeper, id string, sender sdk.AccAddress) sdk.Error {
	return ask.DeleteOrder(ctx, types.MsgCancelOrder{
		Sender:  sender,
		OrderID: id,
	})
}

func removeOrder(t *testing.T,
	ask *autoswap.Keeper, ctx sdk.Context, id string, sender sdk.AccAddress) {
	err := removeOrderWithoutCheck(t, ctx, ask, id, sender)
	require.NoError(t, err)
}

// TODO
func makePrice32(s, e int64) sdk.Dec {
	// s * 10^(e - 23)
	d := sdk.NewDecWithPrec(s, 23-e)
	//println(d.String())
	return d
}

func TestMakePrice32(t *testing.T) {
	require.Equal(t, sdk.MustNewDecFromStr("100.0"), makePrice32(10000000, 18))
	require.Equal(t, sdk.MustNewDecFromStr("10.0"), makePrice32(10000000, 17))
	require.Equal(t, sdk.MustNewDecFromStr("1.0"), makePrice32(10000000, 16))
}
