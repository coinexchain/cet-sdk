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

func newTestApp() (app *testapp.TestApp, ctx sdk.Context) {
	app = testapp.NewTestApp()
	ctx = sdk.NewContext(app.Cms, abci.Header{}, false, log.NewNopLogger())
	app.SupplyKeeper.SetSupply(ctx, supply.Supply{Total: sdk.Coins{}})
	app.AssetKeeper.SetParams(ctx, asset.DefaultParams())
	return
}

func issueToken(t *testing.T, ak asset.Keeper, ctx sdk.Context,
	sym string, totalSupply sdk.Int, owner sdk.AccAddress) {

	err := ak.IssueToken(ctx, sym, sym, totalSupply, owner, false, false, false, false, sym, sym, sym)
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
