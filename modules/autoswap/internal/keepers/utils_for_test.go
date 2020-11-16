package keepers_test

import (
	dex "github.com/coinexchain/cet-sdk/types"
	"testing"

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
	firstSellID  int
}
type PairBooked struct {
	bookedStock int
	bookedMoney int
	firstBuyID  int
}

func (p Pair) mint(stockIn, moneyIn int64, to sdk.AccAddress) {
	mint(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, p.sym, sdk.NewInt(stockIn), sdk.NewInt(moneyIn), to)
}
func (p Pair) addLimitOrder(isBuy bool, sender sdk.AccAddress, amt int64, price sdk.Dec, id int64) {
	addLimitOrder(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, p.sym, isBuy, sender, amt, price, id)
}
func (p Pair) addMarketOrder(isBuy bool, sender sdk.AccAddress, amt int64) {
	addMarketOrder(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, p.sym, isBuy, sender, amt)
}
func (p Pair) removeOrder(isBuy bool, id int64, sender sdk.AccAddress) {
	removeOrder(p.th.t, p.th.app.AutoSwapKeeper, p.th.ctx, p.sym, isBuy, id, sender)
}
func (p Pair) getLiquidity(addr sdk.AccAddress) sdk.Int {
	return p.th.app.AutoSwapKeeper.GetLiquidity(p.th.ctx, p.sym, addr)
}
func (p Pair) getPoolInfo() *autoswap.PoolInfo {
	return p.th.app.AutoSwapKeeper.GetPoolInfo(p.th.ctx, p.sym)
}

func (p Pair) getFirstBuyID() int64 {
	return 0 // TODO
	//return p.th.app.AutoSwapKeeper.GetFirstOrderID(p.th.ctx, p.sym, true)
}
func (p Pair) getFirstSellID() int64 {
	return 0 // TODO
	//return p.th.app.AutoSwapKeeper.GetFirstOrderID(p.th.ctx, p.sym, false)
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
	app.AutoSwapKeeper.SetParams(ctx, autoswap.DefaultParams())
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
	_ sdk.AccAddress, stock, money string, pricePrecision byte) {
	ask.CreatePair(ctx, dex.GetSymbol(stock, money), pricePrecision)
}

func mint(t *testing.T, ask *autoswap.Keeper, ctx sdk.Context,
	pair string, stockIn, moneyIn sdk.Int, to sdk.AccAddress) {

	_, err := ask.Mint(ctx, pair, stockIn, moneyIn, to)
	require.NoError(t, err)
}

func addLimitOrder(t *testing.T, ask *autoswap.Keeper, ctx sdk.Context,
	pair string, isBuy bool, sender sdk.AccAddress, amt int64, price sdk.Dec, id int64) {

	err := ask.AddLimitOrder(ctx, &types.Order{
		Sender:      sender,
		TradingPair: pair,
		IsBuy:       isBuy,
		Price:       price,
		Quantity:    amt,
	})
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

func removeOrder(t *testing.T, ask *autoswap.Keeper, ctx sdk.Context,
	pair string, isBuy bool, id int64, sender sdk.AccAddress) {
	err := ask.DeleteOrder(ctx, types.MsgCancelOrder{
		Sender: sender,
		//OrderID: id,
	})
	require.NoError(t, err)
}

// TODO
func makePrice32(s, e int64) sdk.Dec {
	// s * 10^(e - 23)
	return sdk.NewDecWithPrec(s, 23-e)
}

func TestMakePrice32(t *testing.T) {
	require.Equal(t, sdk.MustNewDecFromStr("100.0"),
		makePrice32(10000000, 18))
}
