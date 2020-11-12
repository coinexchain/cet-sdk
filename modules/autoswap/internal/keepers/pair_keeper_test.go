package keepers_test

import (
	"testing"

	"github.com/coinexchain/cet-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	testapp "github.com/coinexchain/cet-sdk/testapp"
)

var (
	sender = testutil.ToAccAddress("sender")
)

//func TestPairKeeper_SetOrder(t *testing.T) {
//	app := testapp.NewTestApp()
//	ctx := app.NewCtx()
//	k := app.AutoSwapKeeper.IPairKeeper.(*keepers.PairKeeper)
//
//	order := types.Order{
//		Price:           sdk.NewDec(100),
//		OrderID:         10,
//		NextOrderID:     11,
//		PrevKey:         [3]int64{1, 2, 3},
//		MinOutputAmount: sdk.NewInt(19),
//		OrderBasic: types.OrderBasic{
//			MarketSymbol: "stock/money",
//			IsBuy:        false,
//			IsLimitOrder: false,
//			Amount:       sdk.NewInt(1999999),
//		},
//	}
//	k.SetOrder(ctx, &order)
//	recordOrder := k.GetOrder(ctx, order.MarketSymbol, order.IsBuy, order.OrderID)
//	require.NotNil(t, recordOrder)
//	require.EqualValues(t, recordOrder.OrderBasic, order.OrderBasic)
//	require.EqualValues(t, recordOrder.OrderID, order.OrderID)
//	require.EqualValues(t, recordOrder.Price, order.Price)
//	require.EqualValues(t, recordOrder.NextOrderID, order.NextOrderID)
//}

func TestPairKeeper_AllocateFeeToValidatorAndPool(t *testing.T) {
	app := testapp.NewTestApp()
	ctx := app.NewCtx()
	app.AutoSwapKeeper.SetParams(ctx, types.DefaultParams())
	keeper := app.AutoSwapKeeper.IPairKeeper.(*keepers.PairKeeper)

	//setup accounts
	app.AccountKeeper.SetAccount(ctx, supply.NewEmptyModuleAccount(types.PoolModuleAcc))
	app.AccountKeeper.SetAccount(ctx, supply.NewEmptyModuleAccount(auth.FeeCollectorName))
	baseAcc := auth.NewBaseAccount(sender, sdk.NewCoins(sdk.NewCoin("money", sdk.NewInt(100)),
		sdk.NewCoin("stock", sdk.NewInt(100))), nil, 0, 0)
	app.AccountKeeper.SetAccount(ctx, baseAcc)

	err := keeper.AllocateFeeToValidatorAndPool(ctx, "money", sdk.NewInt(5), sender)
	require.Nil(t, err)

	err = keeper.AllocateFeeToValidatorAndPool(ctx, "stock", sdk.NewInt(10), sender)
	require.Nil(t, err)

	poolCoins := app.AccountKeeper.GetAccount(ctx, supply.NewModuleAddress(types.PoolModuleAcc)).GetCoins()
	feeCollectorCoins := app.AccountKeeper.GetAccount(ctx, supply.NewModuleAddress(auth.FeeCollectorName)).GetCoins()
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("money", sdk.NewInt(3)), sdk.NewCoin("stock", sdk.NewInt(6))), poolCoins)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("money", sdk.NewInt(2)), sdk.NewCoin("stock", sdk.NewInt(4))), feeCollectorCoins)
}
