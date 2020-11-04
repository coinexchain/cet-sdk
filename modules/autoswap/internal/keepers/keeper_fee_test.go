package keepers_test

import (
	"testing"

	"github.com/coinexchain/cet-sdk/modules/autoswap"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cet-sdk/testapp"

	dex "github.com/coinexchain/cet-sdk/types"
)

var (
	app *testapp.TestApp
	ctx sdk.Context
)

const (
	stock = "eth"
	money = "usdt"
)

func initialization() {
	app = testapp.NewTestApp()
	ctx = sdk.NewContext(app.Cms, abci.Header{}, false, log.NewNopLogger())
	app.AutoSwapKeeper.SetParams(ctx, types.DefaultParams())

}
func TestKeeper_SetGetParams(t *testing.T) {
	initialization()
	param := app.AutoSwapKeeper.GetParams(ctx)
	require.Equal(t, param.String(), types.DefaultParams().String())
}

func TestKeeper_GetFee(t *testing.T) {
	initialization()
	makerFee := app.AutoSwapKeeper.GetMakerFee(ctx)
	require.Equal(t, sdk.NewDecWithPrec(50, 4), makerFee)

	takerFee := app.AutoSwapKeeper.GetTakerFee(ctx)
	require.Equal(t, sdk.NewDecWithPrec(30, 4), takerFee)

	feeToVal := app.AutoSwapKeeper.GetFeeToValidator(ctx)
	require.Equal(t, sdk.NewDecWithPrec(4, 1), feeToVal)
}

type testCase struct {
	Klast          sdk.Int
	Pool           *keepers.PoolInfo
	tokenAllocated sdk.Coins
}

func newStockMoneyCoins(stockAmount int64, moneyAmount int64) sdk.Coins {
	return sdk.NewCoins(
		sdk.NewCoin(stock, sdk.NewInt(stockAmount)),
		sdk.NewCoin(money, sdk.NewInt(moneyAmount)),
	)

}
func TestKeeper_AllocateFeeToValidator(t *testing.T) {
	initialization()
	info := keepers.NewPoolInfo(dex.GetSymbol(stock, money), sdk.NewInt(10000), sdk.NewInt(10000*400), sdk.NewInt(10000))
	poolAcc := auth.NewBaseAccount(
		app.SupplyKeeper.GetModuleAddress(autoswap.PoolModuleAcc),
		newStockMoneyCoins(10000, 10000*400),
		nil,
		0,
		0)

	app.AccountKeeper.SetAccount(ctx, supply.NewModuleAccount(poolAcc, autoswap.PoolModuleAcc))
	app.AccountKeeper.SetAccount(ctx, supply.NewEmptyModuleAccount(auth.FeeCollectorName))

	testCases := []testCase{
		// klast is 1/10 of k
		{
			Klast: sdk.NewInt(4 * 10000 * 10000),
			Pool:  &info,
			// stock = 18*4/(20*6+2*4)*10000 = 5625
			// money = stock * 400 = 2250000
			tokenAllocated: newStockMoneyCoins(5625, 2250000),
		},
		// klast is 0
		{
			Klast: sdk.NewInt(0),
			Pool:  &info,
		},
		// klast is larger than k
		{
			Klast: sdk.NewInt(4000 * 10000 * 10000),
			Pool:  &info,
		},
	}

	for _, c := range testCases {
		oldCoins := app.SupplyKeeper.GetModuleAccount(ctx, auth.FeeCollectorName).GetCoins()
		_ = app.AutoSwapKeeper.AllocateFeeToValidator(ctx, &c.Klast, c.Pool)

		newCoins := app.SupplyKeeper.GetModuleAccount(ctx, auth.FeeCollectorName).GetCoins()
		require.Equal(t, c.tokenAllocated, newCoins.Sub(oldCoins))

	}

}
