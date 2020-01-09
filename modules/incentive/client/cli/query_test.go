package cli

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/coinexchain/cet-sdk/modules/incentive/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/incentive/internal/types"
	"github.com/coinexchain/cosmos-utils/client/cliutil"
)

func TestQueryParamsCmd(t *testing.T) {
	cmdFactory := func() *cobra.Command {
		return GetQueryCmd(nil)
	}

	cliutil.TestQueryCmd(t, cmdFactory, "params",
		fmt.Sprintf("custom/%s/%s", types.ModuleName, keepers.QueryParameters), nil)
}
