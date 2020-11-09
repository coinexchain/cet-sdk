package keepers_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	testapp "github.com/coinexchain/cet-sdk/testapp"
)

func TestPairKeeper_SetOrder(t *testing.T) {
	app := testapp.NewTestApp()
	ctx := app.NewCtx()
	k := app.AutoSwapKeeper.IPairKeeper.(*keepers.PairKeeper)

	order := types.Order{
		Price:           sdk.NewDec(100),
		OrderID:         10,
		NextOrderID:     11,
		PrevKey:         [3]int64{1, 2, 3},
		MinOutputAmount: sdk.NewInt(19),
		OrderBasic: types.OrderBasic{
			MarketSymbol:    "stock/money",
			IsBuy:           false,
			IsLimitOrder:    false,
			Amount:          sdk.NewInt(1999999),
		},
	}
	k.SetOrder(ctx, &order)
	recordOrder := k.GetOrder(ctx, order.MarketSymbol, order.IsBuy, order.OrderID)
	require.NotNil(t, recordOrder)
	require.EqualValues(t, recordOrder.OrderBasic, order.OrderBasic)
	require.EqualValues(t, recordOrder.OrderID, order.OrderID)
	require.EqualValues(t, recordOrder.Price, order.Price)
	require.EqualValues(t, recordOrder.NextOrderID, order.NextOrderID)
}
