package cli

import (
	"fmt"

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
