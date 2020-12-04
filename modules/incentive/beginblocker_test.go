package incentive_test

import (
	"github.com/coinexchain/cet-sdk/modules/asset"
	"github.com/coinexchain/cet-sdk/modules/incentive/internal/keepers"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/coinexchain/cet-sdk/modules/incentive"
	"github.com/coinexchain/cet-sdk/testapp"
	dex "github.com/coinexchain/cet-sdk/types"
)

type TestInput struct {
	ctx    sdk.Context
	cdc    *codec.Codec
	keeper incentive.Keeper
	ak     auth.AccountKeeper
	sk     supply.Keeper
}

func SetupTestInput() TestInput {
	app := testapp.NewTestApp()
	ctx := sdk.NewContext(app.Cms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())
	cet := asset.BaseToken{
		Name:        "cet",
		Symbol:      "cet",
		Mintable:    true,
		Burnable:    true,
		TotalSupply: sdk.NewInt(1000),
	}
	_ = app.AssetKeeper.SetToken(ctx, &cet)
	app.SupplyKeeper.SetSupply(ctx, supply.Supply{Total: sdk.NewCoins(sdk.NewCoin(cet.Symbol, cet.TotalSupply))})
	return TestInput{ctx: ctx, cdc: app.Cdc, keeper: app.IncentiveKeeper, ak: app.AccountKeeper, sk: app.SupplyKeeper}
}

func TestBeginBlockerInvalidCoin(t *testing.T) {
	input := SetupTestInput()
	_ = input.keeper.SetState(input.ctx, incentive.State{HeightAdjustment: 10})
	input.keeper.SetParams(input.ctx, incentive.DefaultParams())

	feeBalanceBefore := input.sk.GetModuleAccount(input.ctx, auth.FeeCollectorName).GetCoins().AmountOf(dex.CET).Int64()
	incentive.BeginBlocker(input.ctx, input.keeper)
	feeBalanceAfter := input.sk.GetModuleAccount(input.ctx, auth.FeeCollectorName).GetCoins().AmountOf(dex.CET).Int64()

	// no coins in pool
	require.Equal(t, int64(0), feeBalanceAfter-feeBalanceBefore)
}

func TestIncentiveCoinsAddress(t *testing.T) {
	require.Equal(t, "coinex1gc5t98jap4zyhmhmyq5af5s7pyv57w5694el97", keepers.PoolAddr.String())
}

func TestIncentiveCoinsAddressInTestNet(t *testing.T) {
	config := sdk.GetConfig()
	testnetAddrPrefix := "cettest"
	config.SetBech32PrefixForAccount(testnetAddrPrefix, testnetAddrPrefix+sdk.PrefixPublic)
	require.Equal(t, "cettest1gc5t98jap4zyhmhmyq5af5s7pyv57w566ewmx0", keepers.PoolAddr.String())
}

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(dex.Bech32MainPrefix, dex.Bech32MainPrefix+sdk.PrefixPublic)
	os.Exit(m.Run())
}
