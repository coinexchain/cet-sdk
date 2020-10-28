package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coinexchain/cosmos-utils/client/cliutil"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
)

var resultMsg cliutil.MsgWithAccAddress
var fromAddr sdk.AccAddress

func cliRunCmdForTest(cdc *codec.Codec, msg cliutil.MsgWithAccAddress) error {
	cliCtx := context.NewCLIContext().WithCodec(cdc)
	senderAddr := cliCtx.GetFromAddress()
	msg.SetAccAddress(senderAddr)
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	resultMsg = msg
	return nil
}

func setup() {
	sdk.GetConfig().SetBech32PrefixForAccount("coinex", "coinexpub")
	cliutil.CliRunCommand = cliRunCmdForTest

	fromAddr, _ = sdk.AccAddressFromHex("01234567890123456789012345678901234abcde")

}

// https://stackoverflow.com/questions/23729790/how-can-i-do-test-setup-using-the-testing-package-in-go
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	//shutdown()
	os.Exit(code)
}

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
