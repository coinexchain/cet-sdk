package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
)

// get the root query command of this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	// Group asset queries under a subcommand
	assQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the asset module",
	}

	assQueryCmd.AddCommand(client.GetCommands(
	//GetCmdQueryParams(types.QuerierRoute, cdc),
	//GetCmdQueryToken(types.QuerierRoute, cdc),
	//GetCmdQueryTokenList(types.QuerierRoute, cdc),
	//GetCmdQueryTokenWhitelist(types.QuerierRoute, cdc),
	//GetCmdQueryTokenForbiddenAddr(types.QuerierRoute, cdc),
	//GetCmdQueryTokenReservedSymbols(types.QuerierRoute, cdc),
	)...)

	return assQueryCmd
}
