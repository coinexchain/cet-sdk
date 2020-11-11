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
		Sender:  fromAddr,
		To:      fromAddr,
		Stock:   "foo",
		Money:   "bar",
		StockIn: sdk.NewInt(100000000),
		MoneyIn: sdk.NewInt(200000000),
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
		Sender: fromAddr,
		To:     fromAddr,
		Stock:  "foo",
		Money:  "bar",
		Amount: sdk.NewInt(12345),
	}, resultMsg)
}
