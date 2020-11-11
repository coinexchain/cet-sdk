package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
)

func TestAddLiquidityReq(t *testing.T) {
	req := addLiquidityReq{
		Stock:       "foo",
		Money:       "bar",
		NoSwap:      false,
		NoOrderBook: false,
		StockIn:     "123",
		MoneyIn:     "456",
		To:          addr.String(),
	}
	msg, err := req.GetMsg(nil, addr)
	assert.NoError(t, err)
	assert.Equal(t, &types.MsgAddLiquidity{
		Sender:  addr,
		Stock:   "foo",
		Money:   "bar",
		StockIn: sdk.NewInt(123),
		MoneyIn: sdk.NewInt(456),
		To:      addr,
	}, msg)
}

func TestRemoveLiquidityReq(t *testing.T) {
	req := removeLiquidityReq{
		Stock:       "foo",
		Money:       "bar",
		NoSwap:      true,
		NoOrderBook: false,
		StockMin:    "123",
		MoneyMin:    "456",
		Amount:      "789",
		To:          addr.String(),
	}
	msg, err := req.GetMsg(nil, addr)
	assert.NoError(t, err)
	assert.Equal(t, &types.MsgRemoveLiquidity{
		Sender: addr,
		Stock:  "foo",
		Money:  "bar",
		Amount: sdk.NewInt(789),
		To:     addr,
	}, msg)
}

func TestSwapTokensReq(t *testing.T) {
	req := swapTokensReq{
		Path: []pairInfo{
			{
				Symbol:      "foo/bar",
				NoSwap:      false,
				NoOrderBook: false,
			},
		},
		Side:      "buy",
		Amount:    "123",
		OutputMin: "456",
	}
	msg, err := req.GetMsg(nil, addr)
	assert.NoError(t, err)
	assert.Equal(t, &types.MsgSwapTokens{
		Pairs: []types.MarketInfo{
			{
				MarketSymbol: "foo/bar",
			},
		},
		Sender:          addr,
		IsBuy:           true,
		Amount:          sdk.NewInt(123),
		MinOutputAmount: sdk.NewInt(456),
	}, msg)
}

func TestCreateLimitOrderReq(t *testing.T) {
	req := createLimitOrderReq{
		PairSymbol:  "foo/bar",
		NoSwap:      true,
		NoOrderBook: false,
		Side:        "sell",
		Amount:      "123",
		OrderID:     "888",
		Price:       "999.9",
	}
	msg, err := req.GetMsg(nil, addr)
	assert.NoError(t, err)
	assert.Equal(t, &types.MsgCreateLimitOrder{
		OrderBasic: types.OrderBasic{
			Sender:       addr,
			MarketSymbol: "foo/bar",
			IsBuy:        false,
			IsLimitOrder: true,
			Amount:       sdk.NewInt(123),
		},
		OrderID: 888,
		Price:   sdk.MustNewDecFromStr("999.9"),
	}, msg)
}

func TestCancelOrderReq(t *testing.T) {
	req := cancelOrderReq{
		PairSymbol:  "foo/bar",
		NoSwap:      true,
		NoOrderBook: false,
		Side:        "sell",
		OrderID:     "888",
	}
	msg, err := req.GetMsg(nil, addr)
	assert.NoError(t, err)
	assert.Equal(t, &types.MsgDeleteOrder{
		Sender:       addr,
		MarketSymbol: "foo/bar",
		IsBuy:        false,
		OrderID:      888,
	}, msg)
}
