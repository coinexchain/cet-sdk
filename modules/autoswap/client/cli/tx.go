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
	mktcli "github.com/coinexchain/cet-sdk/modules/market/client/cli"
	"github.com/coinexchain/cosmos-utils/client/cliutil"
)

const (
	flagStock   = "stock"
	flagMoney   = "money"
	flagStockIn = "stock-in"
	flagMoneyIn = "money-in"
	flagTo      = "to"
	flagAmount  = "amount"
)

// get the root tx command of this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := mktcli.GetTxCmd(cdc)

	// add new commands
	txCmd.AddCommand(client.PostCommands(
		GetAddLiquidityCmd(cdc),
		GetRemoveLiquidityCmd(cdc),
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
$ cetcli tx autoswap add-liquidity --stock="foo" --money="bar" \
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
$ cetcli tx autoswap remove-liquidity --stock="foo" --money="bar" \
	--amount=12345 \
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
	cmd.Flags().String(flagAmount, "", "the amount of liquidity to be removed")
	cmd.Flags().String(flagTo, "", "mint to")
	markRequiredFlags(cmd, flagStock, flagMoney, flagAmount, flagTo)

	return cmd
}

func addBasicPairFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagStock, "", "the stock symbol of the pool")
	cmd.Flags().String(flagMoney, "", "the money symbol of the pool")
}

func getAddLiquidityMsg() (msg *types.MsgAddLiquidity, err error) {
	msg = &types.MsgAddLiquidity{
		Stock: viper.GetString(flagStock),
		Money: viper.GetString(flagMoney),
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
		Stock: viper.GetString(flagStock),
		Money: viper.GetString(flagMoney),
	}
	if msg.Amount, err = parseSdkInt(flagAmount); err != nil {
		return
	}
	if msg.To, err = sdk.AccAddressFromBech32(viper.GetString(flagTo)); err != nil {
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
