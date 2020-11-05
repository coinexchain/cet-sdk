package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cosmos-utils/client/cliutil"
)

func TestAddLiquidityCmd(t *testing.T) {
	txCmd := GetTxCmd(nil)
	args := []string{
		"add-liquidity",
		"--stock=foo",
		"--money=bar",
		"--no-swap=false",
		"--no-order-book=false",
		"--stock-in=100000000",
		"--money-in=200000000",
		"--to=" + fromAddr.String(),
		"--from=" + fromAddr.String(),
		"--generate-only",
	}
	txCmd.SetArgs(args)
	cliutil.SetViperWithArgs(args)
	err := txCmd.Execute()
	assert.Equal(t, nil, err)
	assert.Equal(t, &types.MsgAddLiquidity{
		Owner:           fromAddr,
		To:              fromAddr,
		Stock:           "foo",
		Money:           "bar",
		StockIn:         sdk.NewInt(100000000),
		MoneyIn:         sdk.NewInt(200000000),
		IsSwapOpen:      true,
		IsOrderBookOpen: true,
	}, resultMsg)
}

func TestRemoveLiquidityCmd(t *testing.T) {
	txCmd := GetTxCmd(nil)
	args := []string{
		"remove-liquidity",
		"--stock=foo",
		"--money=bar",
		"--no-swap=false",
		"--no-order-book=false",
		"--stock-min=100000000",
		"--money-min=200000000",
		"--amount=12345",
		"--to=" + fromAddr.String(),
		"--from=" + fromAddr.String(),
		"--generate-only",
	}
	txCmd.SetArgs(args)
	cliutil.SetViperWithArgs(args)
	err := txCmd.Execute()
	assert.Equal(t, nil, err)
	assert.Equal(t, &types.MsgRemoveLiquidity{
		Sender:          fromAddr,
		To:              fromAddr,
		Stock:           "foo",
		Money:           "bar",
		AmountStockMin:  sdk.NewInt(100000000),
		AmountMoneyMin:  sdk.NewInt(200000000),
		Amount:          sdk.NewInt(12345),
		IsSwapOpen:      true,
		IsOrderBookOpen: true,
	}, resultMsg)
}

func TestSwapTokensCmd(t *testing.T) {
	txCmd := GetTxCmd(nil)
	for _, x := range []struct {
		side  string
		isBuy bool
	}{
		{"buy", true},
		{"sell", false},
	} {
		args := []string{
			"swap-tokens",
			"--swap-path=[{\"pair\":\"foo/bar\", \"noSwap\":false, \"noOrderBook\":false}]",
			"--side=" + x.side,
			"--amount=12345",
			"--output-min=54321",
			"--from=" + fromAddr.String(),
			"--generate-only",
		}
		txCmd.SetArgs(args)
		cliutil.SetViperWithArgs(args)
		err := txCmd.Execute()
		assert.Equal(t, nil, err)
		assert.Equal(t, &types.MsgSwapTokens{
			Pairs: []types.MarketInfo{
				{
					MarketSymbol:    "foo/bar",
					IsOpenSwap:      true,
					IsOpenOrderBook: true,
				},
			},
			Sender:          fromAddr,
			IsBuy:           x.isBuy,
			IsLimitOrder:    false,
			Amount:          sdk.NewInt(12345),
			MinOutputAmount: sdk.NewInt(54321),
		}, resultMsg)
	}
}

func TestCreateLimitOrderCmd(t *testing.T) {
	txCmd := GetTxCmd(nil)
	args := []string{
		"create-limit-order",
		"--pair=foo/bar",
		"--no-swap",
		//"--no-order-book",
		"--side=buy",
		"--amount=12345",
		"--price=678.9",
		"--order-id=6789",
		"--from=" + fromAddr.String(),
		"--generate-only",
	}
	txCmd.SetArgs(args)
	cliutil.SetViperWithArgs(args)
	err := txCmd.Execute()
	assert.Equal(t, nil, err)
	assert.Equal(t, &types.MsgCreateLimitOrder{
		OrderBasic: types.OrderBasic{
			Sender:          fromAddr,
			MarketSymbol:    "foo/bar",
			IsOpenSwap:      false,
			IsOpenOrderBook: true,
			IsBuy:           true,
			IsLimitOrder:    true,
			Amount:          sdk.NewInt(12345),
		},
		Price:   sdk.MustNewDecFromStr("678.9"),
		OrderID: 6789,
	}, resultMsg)
}

func TestDeleteOrderCmd(t *testing.T) {
	txCmd := GetTxCmd(nil)
	args := []string{
		"delete-order",
		"--pair=foo/bar",
		"--no-swap",
		//"--no-order-book",
		"--side=buy",
		"--order-id=6789",
		"--from=" + fromAddr.String(),
		"--generate-only",
	}
	txCmd.SetArgs(args)
	cliutil.SetViperWithArgs(args)
	err := txCmd.Execute()
	assert.Equal(t, nil, err)
	assert.Equal(t, &types.MsgDeleteOrder{
		Sender:          fromAddr,
		MarketSymbol:    "foo/bar",
		IsOpenSwap:      false,
		IsOpenOrderBook: true,
		IsBuy:           true,
		OrderID:         6789,
	}, resultMsg)
}
