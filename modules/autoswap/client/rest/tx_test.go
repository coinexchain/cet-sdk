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
