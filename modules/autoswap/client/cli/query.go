package cli

import (
	"fmt"
	dex "github.com/coinexchain/cet-sdk/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cosmos-utils/client/cliutil"
)

// get the root query command of this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	// Group asset queries under a subcommand
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the asset module",
	}

	queryCmd.AddCommand(client.GetCommands(
		GetQueryParamsCmd(cdc),
		GetQueryPoolCmd(cdc),
		GetQueryPoolListCmd(cdc),
		//GetCmdQueryToken(types.QuerierRoute, cdc),
		//GetCmdQueryTokenList(types.QuerierRoute, cdc),
		//GetCmdQueryTokenWhitelist(types.QuerierRoute, cdc),
		//GetCmdQueryTokenForbiddenAddr(types.QuerierRoute, cdc),
		//GetCmdQueryTokenReservedSymbols(types.QuerierRoute, cdc),
	)...)

	return queryCmd
}

func GetQueryParamsCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query autoswap params",
		RunE: func(cmd *cobra.Command, args []string) error {
			route := fmt.Sprintf("custom/%s/%s", types.StoreKey, keepers.QueryParameters)
			return cliutil.CliQuery(cdc, route, nil)
		},
	}
}

func GetQueryPoolCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pool-info [stock] [money]",
		Args:  cobra.ExactArgs(2),
		Short: "Query autoswap pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			route := fmt.Sprintf("custom/%s/%s", types.StoreKey, keepers.QueryPoolInfo)
			symbol := dex.GetSymbol(args[0], args[1])
			p := keepers.QueryPoolInfoParam{Symbol: symbol}
			return cliutil.CliQuery(cdc, route, p)
		},
	}
}

func GetQueryPoolListCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pool-list",
		Args:  cobra.NoArgs,
		Short: "Query autoswap pool list",
		RunE: func(cmd *cobra.Command, args []string) error {
			route := fmt.Sprintf("custom/%s/%s", types.StoreKey, keepers.QueryPools)
			return cliutil.CliQuery(cdc, route, nil)
		},
	}
}
