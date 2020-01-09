package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/coinexchain/cet-sdk/modules/incentive/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/incentive/internal/types"
	"github.com/coinexchain/cosmos-utils/client/restutil"
)

func TestQueryParamsHandlerFn(t *testing.T) {

	expectedQuery := fmt.Sprintf("custom/%s/%s", types.ModuleName, keepers.QueryParameters)
	oldRestQuery := restutil.RestQuery
	executed := false
	restutil.RestQuery = func(cdc *codec.Codec, cliCtx context.CLIContext, w http.ResponseWriter, r *http.Request, query string, param interface{}, defaultRes []byte) {
		require.Equal(t, expectedQuery, query)
		executed = true
	}

	defer func() {
		restutil.RestQuery = oldRestQuery
	}()
	queryParamsHandlerFn(context.NewCLIContextWithFrom(""))(nil, nil)
	require.True(t, executed)
}
