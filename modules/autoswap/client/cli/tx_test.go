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
		"--no-swap",
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
		Owner:      fromAddr,
		To:         fromAddr,
		Stock:      "foo",
		Money:      "bar",
		StockIn:    sdk.NewInt(100000000),
		MoneyIn:    sdk.NewInt(200000000),
		IsOpenSwap: false,
	}, resultMsg)
}

func TestCreateMarketOrderCmd(t *testing.T) {
	txCmd := GetTxCmd(nil)
	for _, x := range []struct {
		side  string
		isBuy bool
	}{
		{"buy", true},
		{"sell", false},
	} {
		args := []string{
			"create-market-order",
			"--pair=foo/bar",
			"--side=" + x.side,
			"--amount=12345",
			"--no-swap",
			"--from=" + fromAddr.String(),
			"--generate-only",
		}
		txCmd.SetArgs(args)
		cliutil.SetViperWithArgs(args)
		err := txCmd.Execute()
		assert.Equal(t, nil, err)
		assert.Equal(t, &types.MsgCreateMarketOrder{
			OrderBasic: types.OrderBasic{
				Sender:       fromAddr,
				MarketSymbol: "foo/bar",
				Amount:       12345,
				IsBuy:        x.isBuy,
				IsOpenSwap:   false,
			},
		}, resultMsg)
	}
}

func TestGetCreateLimitOrderCmd(t *testing.T) {
	txCmd := GetTxCmd(nil)
	args := []string{
		"create-limit-order",
		"--pair=foo/bar",
		"--side=buy",
		"--amount=12345",
		"--no-swap",
		"--price=10000",
		"--price-precision=8",
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
			Sender:       fromAddr,
			MarketSymbol: "foo/bar",
			Amount:       12345,
			IsBuy:        true,
			IsOpenSwap:   false,
		},
		Price:          10000,
		PricePrecision: 8,
		OrderID:        6789,
	}, resultMsg)
}
