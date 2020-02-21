package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/coinexchain/cosmos-utils/client/restutil"
)

func setRefereeHandleFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return restutil.NewRestHandlerBuilder(cdc, cliCtx, new(setRefereeReq)).Build(nil)
}
