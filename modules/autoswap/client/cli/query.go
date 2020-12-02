package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/market"
	mktcli "github.com/coinexchain/cet-sdk/modules/market/client/cli"
	dex "github.com/coinexchain/cet-sdk/types"
	"github.com/coinexchain/cosmos-utils/client/cliutil"
)

// get the root query command of this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	queryCmd := mktcli.GetQueryCmd(cdc)
	// TODO: remove unsupported commands

	// add new commands
	queryCmd.AddCommand(client.GetCommands(
		GetQueryPoolCmd(cdc),
		GetQueryPoolListCmd(cdc),
	)...)

	return queryCmd
}

func GetQueryPoolCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pool-info [stock] [money]",
		Args:  cobra.ExactArgs(2),
		Short: "Query market pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			route := fmt.Sprintf("custom/%s/%s", market.StoreKey, keepers.QueryPoolInfo)
			symbol := dex.GetSymbol(args[0], args[1])
			p := market.QueryMarketParam{TradingPair: symbol}
			return cliutil.CliQuery(cdc, route, p)
		},
	}
}

func GetQueryPoolListCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pool-list",
		Args:  cobra.NoArgs,
		Short: "Query market pool list",
		RunE: func(cmd *cobra.Command, args []string) error {
			route := fmt.Sprintf("custom/%s/%s", market.StoreKey, keepers.QueryPools)
			return cliutil.CliQuery(cdc, route, nil)
		},
	}
}
