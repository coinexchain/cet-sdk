package keepers

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	QueryParameters = "parameters"
	QueryPoolInfo   = "pool-info"
	QueryPools      = "pool-list"
)

// creates a querier for asset REST endpoints
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryParameters:
			return queryParameters(ctx, keeper)
		case QueryPoolInfo:
			return queryPoolInfo(ctx, req, keeper)
		case QueryPools:
			return queryPoolList(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("query Symbol : " + path[0])
		}
	}
}

func queryParameters(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

type QueryPoolInfoParam struct {
	Symbol        string `json:"Symbol"`
	IsSwapOpen       bool   `json:"amm_open"`
	OrderBookOpen bool   `json:"order_book_open"`
}

func queryPoolInfo(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var param QueryPoolInfoParam
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, sdk.NewError(types.CodeSpaceAutoSwap, types.CodeUnMarshalFailed, "failed to parse param")
	}
	info := keeper.IPairKeeper.GetPoolInfo(ctx, param.Symbol, true, false)
	var infoD PoolInfoDisplay
	if info != nil {
		infoD = NewPoolInfoDisplay(info)
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, infoD)
	if err != nil {
		return nil, types.ErrMarshalFailed()
	}
	return bz, nil
}

func queryPoolList(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	return nil, nil
}
