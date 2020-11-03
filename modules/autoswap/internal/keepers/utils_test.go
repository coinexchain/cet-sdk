package keepers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
)

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
