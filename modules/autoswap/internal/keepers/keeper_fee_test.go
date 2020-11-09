package keepers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cet-sdk/testapp"
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
