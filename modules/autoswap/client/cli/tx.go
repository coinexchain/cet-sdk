package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cosmos-utils/client/cliutil"
)

const (
	flagStock          = "stock"
	flagMoney          = "money"
	flagInitStock      = "init-stock"
	flagInitMoney      = "init-money"
	flagNoSwap         = "no-swap"
	flagTo             = "to"
	flagPair           = "pair"
	flagAmount         = "amount"
	flagSide           = "side"
	flagOrderID        = "order-id"
	flagPrice          = "price"
	flagPricePrecision = "price-precision"
	flagPrevKey        = "prev-key"
)

// get the root tx command of this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	assTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Asset transactions subcommands",
	}

	assTxCmd.AddCommand(client.PostCommands(
		GetCreatePairCmd(cdc),
		GetCreateLimitOrderCmd(cdc),
		GetCreateMarketOrderCmd(cdc),
	)...)

	return assTxCmd
}

func GetCreatePairCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pair",
		Short: "generate tx to create autoswap pair",
		Long: strings.TrimSpace(
			`generate a tx and sign it to create autoswap pair in Dex blockchain. 

Example:
$ cetcli tx autoswap create-pair --stock="foo" --money="bar" \
	--init-stock=100000000 --init-money=100000000 \
	--no-swap --to=coinex1px8alypku5j84qlwzdpynhn4nyrkagaytu5u4a \
	--from=bob --chain-id=coinexdex --gas=1000000 --fees=1000cet
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := getCreatePairMsg()
			if err != nil {
				return err
			}
			return cliutil.CliRunCommand(cdc, msg)
		},
	}

	cmd.Flags().String(flagStock, "", "the stock symbol of the pool")
	cmd.Flags().String(flagMoney, "", "the money symbol of the pool")
	cmd.Flags().String(flagInitStock, "", "the init stock amount of the pool")
	cmd.Flags().String(flagInitMoney, "", "the init money amount of the pool")
	cmd.Flags().Bool(flagNoSwap, false, "disable swap function")
	cmd.Flags().String(flagTo, "", "mint to")
	markRequiredFlags(cmd, flagStock, flagMoney, flagInitStock, flagInitMoney, flagTo)

	return cmd
}

func GetCreateMarketOrderCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-market-order",
		Short: "generate tx to create autoswap market order",
		Long: strings.TrimSpace(
			`generate a tx and sign it to create autoswap market order in Dex blockchain. 

Example:
$ cetcli tx autoswap create-market-order --pair="foo/bar" --no-swap \
	--side=buy --amount=12345 \
	--from=bob --chain-id=coinexdex --gas=1000000 --fees=1000cet
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := getCreateMarketOrderMsg()
			if err != nil {
				return err
			}
			return cliutil.CliRunCommand(cdc, msg)
		},
	}

	cmd.Flags().String(flagPair, "", "the symbol of the autoswap pair")
	cmd.Flags().Bool(flagNoSwap, false, "whether this pool support swap function")
	cmd.Flags().String(flagSide, "", "buy or sell")
	cmd.Flags().Int64(flagAmount, 0, "the amount of the order")
	markRequiredFlags(cmd, flagPair, flagSide, flagAmount)

	return cmd
}

func GetCreateLimitOrderCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-limit-order",
		Short: "generate tx to create autoswap limit order",
		Long: strings.TrimSpace(
			`generate a tx and sign it to create autoswap limit order in Dex blockchain. 

Example:
$ cetcli tx autoswap create-limit-order --pool="foo/bar" --no-swap \
	--side=buy --amount=12345 \
	--price=10000 --price-precision=8 --order-id=123 --prev-key="4,5,6"
	--from=bob --chain-id=coinexdex --gas=1000000 --fees=1000cet
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := getCreateLimitOrderMsg()
			if err != nil {
				return err
			}
			return cliutil.CliRunCommand(cdc, msg)
		},
	}

	cmd.Flags().String(flagPair, "", "the symbol of the autoswap pair")
	cmd.Flags().Bool(flagNoSwap, false, "whether this pool support swap function")
	cmd.Flags().String(flagSide, "", "buy or sell")
	cmd.Flags().Int64(flagAmount, 0, "the amount of the order")
	cmd.Flags().Uint64(flagPrice, 0, "the price of the order")
	cmd.Flags().Uint8(flagPricePrecision, 0, "the price precision of the order")
	cmd.Flags().Uint64(flagOrderID, 0, "the order ID")
	cmd.Flags().String(flagPrevKey, "", "previous keys, at most 3, separated by comma")

	markRequiredFlags(cmd, flagPair, flagSide, flagAmount,
		flagPrice, flagPricePrecision, flagOrderID)

	return cmd
}

func getCreatePairMsg() (msg *types.MsgCreatePair, err error) {
	msg = &types.MsgCreatePair{
		Stock:      viper.GetString(flagStock),
		Money:      viper.GetString(flagMoney),
		IsOpenSwap: !viper.GetBool(flagNoSwap),
	}

	if msg.StockIn, err = parseSdkInt(flagInitStock); err != nil {
		return
	}
	if msg.MoneyIn, err = parseSdkInt(flagInitMoney); err != nil {
		return
	}
	if msg.To, err = sdk.AccAddressFromBech32(viper.GetString(flagTo)); err != nil {
		return
	}
	return
}

func getCreateMarketOrderMsg() (msg *types.MsgCreateMarketOrder, err error) {
	msg = &types.MsgCreateMarketOrder{}
	msg.IsLimitOrder = false
	if msg.OrderBasic, err = getOrderBasic(); err != nil {
		return
	}
	if msg.IsBuy, err = parseIsBuy(); err != nil {
		return
	}
	return
}

func getCreateLimitOrderMsg() (msg *types.MsgCreateLimitOrder, err error) {
	msg = &types.MsgCreateLimitOrder{}
	msg.IsLimitOrder = true
	if msg.OrderBasic, err = getOrderBasic(); err != nil {
		return
	}
	msg.OrderID = viper.GetUint64(flagOrderID)
	msg.Price = viper.GetUint64(flagPrice)
	msg.PricePrecision = uint8(viper.GetUint(flagPricePrecision))
	msg.PrevKey = [3]int64{} // TODO
	return
}

func getOrderBasic() (ob types.OrderBasic, err error) {
	ob.MarketSymbol = viper.GetString(flagPair)
	ob.IsOpenSwap = !viper.GetBool(flagNoSwap)
	ob.Amount = viper.GetInt64(flagAmount)
	ob.IsBuy, err = parseIsBuy()
	return
}

func markRequiredFlags(cmd *cobra.Command, flagNames ...string) {
	for _, flagName := range flagNames {
		cmd.MarkFlagRequired(flagName)
	}
}
func parseSdkInt(flagName string) (val sdk.Int, err error) {
	flagVal := viper.GetString(flagName)

	ok := false
	if val, ok = sdk.NewIntFromString(flagVal); !ok {
		err = fmt.Errorf("%s must be a valid integer", flagName)
	}
	return
}
func parseIsBuy() (bool, error) {
	side := strings.ToLower(viper.GetString(flagSide))
	if side == "buy" {
		return true, nil
	}
	if side == "sell" {
		return false, nil
	}
	return false, fmt.Errorf("%s must be buy or sell", flagSide)
}
