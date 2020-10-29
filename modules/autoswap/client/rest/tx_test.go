package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
)

func TestAddLiquidityReq(t *testing.T) {
	req := addLiquidityReq{
		Stock:   "foo",
		Money:   "bar",
		StockIn: "123",
		MoneyIn: "456",
		NoSwap:  true,
		To:      addr.String(),
	}
	msg, err := req.GetMsg(nil, addr)
	assert.NoError(t, err)
	assert.Equal(t, &types.MsgAddLiquidity{
		Owner:      addr,
		Stock:      "foo",
		Money:      "bar",
		StockIn:    sdk.NewInt(123),
		MoneyIn:    sdk.NewInt(456),
		IsOpenSwap: false,
		To:         addr,
	}, msg)
}

func TestCreateMarketOrderReq(t *testing.T) {
	req := createMarketOrderReq{
		PairSymbol: "foo/bar",
		NoSwap:     true,
		Side:       "buy",
		Amount:     "123",
	}
	msg, err := req.GetMsg(nil, addr)
	assert.NoError(t, err)
	assert.Equal(t, &types.MsgCreateMarketOrder{
		OrderBasic: types.OrderBasic{
			Sender:       addr,
			MarketSymbol: "foo/bar",
			IsOpenSwap:   false,
			IsBuy:        true,
			IsLimitOrder: false,
			Amount:       123,
		},
	}, msg)
}

func TestCreateLimitOrderReq(t *testing.T) {
	req := createLimitOrderReq{
		PairSymbol:     "foo/bar",
		NoSwap:         true,
		Side:           "sell",
		Amount:         "123",
		OrderID:        "321",
		Price:          "999",
		PricePrecision: 8,
	}
	msg, err := req.GetMsg(nil, addr)
	assert.NoError(t, err)
	assert.Equal(t, &types.MsgCreateLimitOrder{
		OrderBasic: types.OrderBasic{
			Sender:       addr,
			MarketSymbol: "foo/bar",
			IsOpenSwap:   false,
			IsBuy:        false,
			IsLimitOrder: true,
			Amount:       123,
		},
		OrderID:        321,
		Price:          999,
		PricePrecision: 8,
	}, msg)
}
