package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/coinexchain/cet-sdk/modules/market"
	mktcli "github.com/coinexchain/cet-sdk/modules/market/client/cli"
)

// get the root query command of this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	mktCmd := getMarketQueryCmd(cdc)
	// TODO: remove unsupported commands
	return mktCmd
}

func getMarketQueryCmd(cdc *codec.Codec) *cobra.Command {
	mktTxCmd := &cobra.Command{
		Use:   market.ModuleName,
		Short: "Querying commands for the market module",
	}
	mktTxCmd.AddCommand(client.PostCommands(
		mktcli.QueryParamsCmd(cdc),
		mktcli.QueryMarketCmd(cdc),
		mktcli.QueryMarketListCmd(cdc),
		mktcli.QueryOrderbookCmd(cdc),
		mktcli.QueryOrderCmd(cdc),
		mktcli.QueryUserOrderList(cdc),
	)...)
	return mktTxCmd
}
