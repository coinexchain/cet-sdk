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
	flagStock       = "stock"
	flagMoney       = "money"
	flagNoSwap      = "no-swap"
	flagNoOrderBook = "no-order-book"
	flagStockIn     = "stock-in"
	flagMoneyIn     = "money-in"
	flagStockMin    = "stock-min"
	flagMoneyMin    = "money-min"
	flagOutputMin   = "output-min"
	flagTo          = "to"
	flagPairSymbol  = "pair"
	flagAmount      = "amount"
	flagSide        = "side"
	flagOrderID     = "order-id"
	flagPrice       = "price"
	flagPrevKey     = "prev-key"
)

// get the root tx command of this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Asset transactions subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		GetAddLiquidityCmd(cdc),
		GetRemoveLiquidityCmd(cdc),
		GetCreateLimitOrderCmd(cdc),
		GetCreateMarketOrderCmd(cdc),
		GetDeleteOrderCmd(cdc),
	)...)

	return txCmd
}

func GetAddLiquidityCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity",
		Short: "generate tx to add liquidity into autoswap pair",
		Long: strings.TrimSpace(
			`generate a tx and sign it to add liquidity into autoswap pair in Dex blockchain. 

Example:
$ cetcli tx autoswap add-liquidity --stock="foo" --money="bar" --no-swap \
	--stock-in=100000000 --money-in=100000000 \
	--to=coinex1px8alypku5j84qlwzdpynhn4nyrkagaytu5u4a \
	--from=bob --chain-id=coinexdex --gas=1000000 --fees=1000cet
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := getAddLiquidityMsg()
			if err != nil {
				return err
			}
			return cliutil.CliRunCommand(cdc, msg)
		},
	}

	addBasicPairFlags(cmd)
	cmd.Flags().String(flagStockIn, "", "the amount of stock to put into the pool")
	cmd.Flags().String(flagMoneyIn, "", "the amount of money to put into the pool")
	cmd.Flags().String(flagTo, "", "mint to")
	markRequiredFlags(cmd, flagStock, flagMoney,
		flagStockIn, flagMoneyIn, flagTo)

	return cmd
}

func GetRemoveLiquidityCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity",
		Short: "generate tx to remove liquidity from autoswap pair",
		Long: strings.TrimSpace(
			`generate a tx and sign it to remove liquidity from autoswap pair in Dex blockchain. 

Example:
$ cetcli tx autoswap remove-liquidity --stock="foo" --money="bar" --no-swap \
	--stock-in=100000000 --money-in=100000000 \
	--to=coinex1px8alypku5j84qlwzdpynhn4nyrkagaytu5u4a \
	--from=bob --chain-id=coinexdex --gas=1000000 --fees=1000cet
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := getRemoveLiquidityMsg()
			if err != nil {
				return err
			}
			return cliutil.CliRunCommand(cdc, msg)
		},
	}

	addBasicPairFlags(cmd)
	cmd.Flags().String(flagStockMin, "", "the minimum amount of got stock")
	cmd.Flags().String(flagMoneyMin, "", "the minimum amount of got money")
	cmd.Flags().String(flagAmount, "", "the amount of liquidity to be removed")
	cmd.Flags().String(flagTo, "", "mint to")
	markRequiredFlags(cmd, flagStock, flagMoney,
		flagStockMin, flagMoneyMin, flagAmount, flagTo)

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

	addBasicOrderFlags(cmd)
	cmd.Flags().String(flagAmount, "", "the amount of the order")
	cmd.Flags().String(flagOutputMin, "", "the minimum output")
	markRequiredFlags(cmd, flagPairSymbol, flagSide, flagAmount)

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

	addBasicOrderFlags(cmd)
	cmd.Flags().String(flagAmount, "", "the amount of the order")
	cmd.Flags().String(flagPrice, "", "the price of the order")
	cmd.Flags().Int64(flagOrderID, 0, "the order ID")
	cmd.Flags().String(flagPrevKey, "", "previous keys, at most 3, separated by comma")

	markRequiredFlags(cmd, flagPairSymbol, flagSide,
		flagAmount, flagPrice, flagOrderID)

	return cmd
}

func GetDeleteOrderCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-order",
		Short: "generate tx to delete autoswap order",
		Long: strings.TrimSpace(
			`generate a tx and sign it to delete autoswap order in Dex blockchain. 

Example:
$ cetcli tx autoswap delete-order --pool="foo/bar" --no-swap \
	--side=buy --amount=12345 \
	--price=10000 --price-precision=8 --order-id=123 --prev-key="4,5,6"
	--from=bob --chain-id=coinexdex --gas=1000000 --fees=1000cet
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := getDeleteOrderMsg()
			if err != nil {
				return err
			}
			return cliutil.CliRunCommand(cdc, msg)
		},
	}

	addBasicOrderFlags(cmd)
	cmd.Flags().Int64(flagOrderID, 0, "the order ID")
	cmd.Flags().String(flagPrevKey, "", "previous keys, at most 3, separated by comma")

	markRequiredFlags(cmd, flagPairSymbol, flagSide, flagOrderID)

	return cmd
}

func addBasicPairFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagStock, "", "the stock symbol of the pool")
	cmd.Flags().String(flagMoney, "", "the money symbol of the pool")
	cmd.Flags().Bool(flagNoSwap, false, "whether swap function is disabled")
	cmd.Flags().Bool(flagNoOrderBook, false, "whether order book is disabled")
}

func addBasicOrderFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagPairSymbol, "", "the symbol of the autoswap pair")
	cmd.Flags().Bool(flagNoSwap, false, "whether swap function is disabled")
	cmd.Flags().Bool(flagNoOrderBook, false, "whether order book is disabled")
	cmd.Flags().String(flagSide, "", "buy or sell")
}

func getAddLiquidityMsg() (msg *types.MsgAddLiquidity, err error) {
	msg = &types.MsgAddLiquidity{
		Stock:      viper.GetString(flagStock),
		Money:      viper.GetString(flagMoney),
		IsSwapOpen: !viper.GetBool(flagNoSwap),
	}

	if msg.StockIn, err = parseSdkInt(flagStockIn); err != nil {
		return
	}
	if msg.MoneyIn, err = parseSdkInt(flagMoneyIn); err != nil {
		return
	}
	if msg.To, err = sdk.AccAddressFromBech32(viper.GetString(flagTo)); err != nil {
		return
	}
	return
}

func getRemoveLiquidityMsg() (msg *types.MsgRemoveLiquidity, err error) {
	msg = &types.MsgRemoveLiquidity{
		Stock:           viper.GetString(flagStock),
		Money:           viper.GetString(flagMoney),
		IsSwapOpen:      !viper.GetBool(flagNoSwap),
		IsOrderBookOpen: !viper.GetBool(flagNoOrderBook),
	}

	if msg.AmountStockMin, err = parseSdkInt(flagStockMin); err != nil {
		return
	}
	if msg.AmountMoneyMin, err = parseSdkInt(flagMoneyMin); err != nil {
		return
	}
	if msg.Amount, err = parseSdkInt(flagAmount); err != nil {
		return
	}
	if msg.To, err = sdk.AccAddressFromBech32(viper.GetString(flagTo)); err != nil {
		return
	}
	return
}

// todo. will adapter MsgSwapTokens later.
func getCreateMarketOrderMsg() (msg *types.MsgSwapTokens, err error) {
	msg = &types.MsgSwapTokens{}
	//if msg.OrderBasic, err = getOrderBasic(); err != nil {
	//	return
	//}
	msg.IsLimitOrder = false
	if msg.MinOutputAmount, err = parseSdkInt(flagOutputMin); err != nil {
		return
	}
	return
}

func getCreateLimitOrderMsg() (msg *types.MsgCreateLimitOrder, err error) {
	msg = &types.MsgCreateLimitOrder{}
	if msg.OrderBasic, err = getOrderBasic(); err != nil {
		return
	}
	msg.IsLimitOrder = true
	if msg.Price, err = parseSdkDec(flagPrice); err != nil {
		return
	}
	msg.OrderID = viper.GetInt64(flagOrderID)
	// TODO: PrevKey
	return
}

func getDeleteOrderMsg() (msg *types.MsgDeleteOrder, err error) {
	msg = &types.MsgDeleteOrder{
		MarketSymbol:    viper.GetString(flagPairSymbol),
		IsOpenSwap:      !viper.GetBool(flagNoSwap),
		IsOpenOrderBook: !viper.GetBool(flagNoOrderBook),
		OrderID:         viper.GetInt64(flagOrderID),
	}
	if msg.IsBuy, err = parseIsBuy(); err != nil {
		return
	}
	// TODO: PrevKey
	return msg, nil
}

func getOrderBasic() (ob types.OrderBasic, err error) {
	ob.MarketSymbol = viper.GetString(flagPairSymbol)
	ob.IsOpenSwap = !viper.GetBool(flagNoSwap)
	ob.IsOpenOrderBook = !viper.GetBool(flagNoOrderBook)
	if ob.Amount, err = parseSdkInt(flagAmount); err != nil {
		return
	}
	if ob.IsBuy, err = parseIsBuy(); err != nil {
		return
	}
	return
}

func markRequiredFlags(cmd *cobra.Command, flagNames ...string) error {
	for _, flagName := range flagNames {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return err
		}
	}
	return nil
}
func parseSdkInt(flagName string) (val sdk.Int, err error) {
	flagVal := viper.GetString(flagName)

	ok := false
	if val, ok = sdk.NewIntFromString(flagVal); !ok {
		err = fmt.Errorf("%s must be a valid integer number", flagName)
	}
	return
}
func parseSdkDec(flagName string) (val sdk.Dec, err error) {
	flagVal := viper.GetString(flagName)
	if val, err = sdk.NewDecFromStr(flagVal); err != nil {
		err = fmt.Errorf("%s must be a valid decimal number", flagName)
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
