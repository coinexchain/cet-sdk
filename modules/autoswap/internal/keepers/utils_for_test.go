package keepers_test

import (
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
func (h TestHelper) issueToken(sym string, totalSupply int64, owner sdk.AccAddress) {
	issueToken(h.t, h.app.AssetKeeper, h.ctx, sym, sdk.NewInt(totalSupply), owner)
}
func (h TestHelper) balanceOf(sym string, addr sdk.AccAddress) sdk.Int {
	return h.app.BankxKeeper.GetCoins(h.ctx, addr).AmountOf(sym)
}
func (h TestHelper) transfer(sym string, amt int64, from, to sdk.AccAddress) {
	err := h.app.BankxKeeper.SendCoins(h.ctx, from, to, sdk.NewCoins(sdk.NewCoin(sym, sdk.NewInt(amt))))
	require.NoError(h.t, err)
}
func (h TestHelper) createPair(owner sdk.AccAddress, stock, money string) {
	createPair(h.t, h.app.AutoSwapKeeper, h.ctx, owner, stock, money)
}
func (h TestHelper) mint(pair string, stockIn, moneyIn int64, to sdk.AccAddress) {
	mint(h.t, h.app.AutoSwapKeeper, h.ctx, pair, sdk.NewInt(stockIn), sdk.NewInt(moneyIn), to)
}
func (h TestHelper) addLimitOrder(pair string, isBuy bool, sender sdk.AccAddress, amt int64, price sdk.Dec, id int64, prevKey [3]int64) {
	addLimitOrder(h.t, h.app.AutoSwapKeeper, h.ctx, pair, isBuy, sender, sdk.NewInt(amt), price, id, prevKey)
}
func (h TestHelper) removeOrder(pair string, isBuy bool, id int64, prevKey [3]int64, sender sdk.AccAddress) {
	removeOrder(h.t, h.app.AutoSwapKeeper, h.ctx, pair, isBuy, id, prevKey, sender)
}
func (h TestHelper) getLiquidity(pair string, addr sdk.AccAddress) sdk.Int {
	return h.app.AutoSwapKeeper.GetLiquidity(h.ctx, pair, true, true, addr)
}
func (h TestHelper) getPoolInfo(pair string) *autoswap.PoolInfo {
	return h.app.AutoSwapKeeper.GetPoolInfo(h.ctx, pair, true, true)
}
func (h TestHelper) getFirstBuyID(pair string) int64 {
	return h.app.AutoSwapKeeper.GetFirstOrderID(h.ctx, pair, true, true, true)
}
func (h TestHelper) getFirstSellID(pair string) int64 {
	return h.app.AutoSwapKeeper.GetFirstOrderID(h.ctx, pair, true, true, false)
}

/* helper functions */

func newTestApp() (app *testapp.TestApp, ctx sdk.Context) {
	app = testapp.NewTestApp()
	ctx = sdk.NewContext(app.Cms, abci.Header{}, false, log.NewNopLogger())
	app.SupplyKeeper.SetSupply(ctx, supply.Supply{Total: sdk.Coins{}})
	app.AssetKeeper.SetParams(ctx, asset.DefaultParams())
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

func createPair(t *testing.T, ask autoswap.Keeper, ctx sdk.Context,
	owner sdk.AccAddress, stock, money string) {

	err := ask.CreatePair(ctx, types.MsgAddLiquidity{
		Owner:           owner,
		Stock:           stock,
		Money:           money,
		IsSwapOpen:      true,
		IsOrderBookOpen: true,
		StockIn:         sdk.NewInt(0),
		MoneyIn:         sdk.NewInt(0),
		To:              nil,
	})
	require.NoError(t, err)
}

func mint(t *testing.T, ask autoswap.Keeper, ctx sdk.Context,
	pair string, stockIn, moneyIn sdk.Int, to sdk.AccAddress) {

	err := ask.Mint(ctx, pair, true, true,
		stockIn, moneyIn, to)
	require.NoError(t, err)
}

func addLimitOrder(t *testing.T, ask autoswap.Keeper, ctx sdk.Context,
	pair string, isBuy bool, sender sdk.AccAddress, amt sdk.Int, price sdk.Dec, id int64, prevKey [3]int64) {

	err := ask.AddLimitOrder(ctx, &types.Order{
		OrderBasic: types.OrderBasic{
			MarketSymbol:    pair,
			IsOpenSwap:      true,
			IsOpenOrderBook: true,
			IsLimitOrder:    true,
			IsBuy:           isBuy,
			Sender:          sender,
			Amount:          amt,
		},
		Price:   price,
		OrderID: id,
		PrevKey: prevKey,
	})
	require.NoError(t, err)
}

func removeOrder(t *testing.T, ask autoswap.Keeper, ctx sdk.Context,
	pair string, isBuy bool, id int64, prevKey [3]int64, sender sdk.AccAddress) {
	err := ask.DeleteOrder(ctx, &types.MsgDeleteOrder{
		MarketSymbol:    pair,
		IsOpenSwap:      true,
		IsOpenOrderBook: true,
		Sender:          sender,
		IsBuy:           isBuy,
		OrderID:         id,
		PrevKey:         prevKey,
	})
	require.NoError(t, err)
}

func merge3(a, b, c int64) [3]int64 {
	return [3]int64{a, b, c}
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
