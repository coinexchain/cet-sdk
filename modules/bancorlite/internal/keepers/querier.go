package keepers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/coinexchain/cet-sdk/modules/bancorlite/internal/types"
)

const (
	QueryBancorInfo = "bancor-info"
	QueryParameters = "parameters"
	QueryBancors    = "bancor-list"
)

// creates a querier for asset REST endpoints
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryParameters:
			return queryParameters(ctx, keeper)
		case QueryBancorInfo:
			return queryBancorInfo(ctx, req, keeper)
		case QueryBancors:
			return queryBancorList(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("query symbol : " + path[0])
		}
	}
}

type QueryBancorInfoParam struct {
	Symbol string `json:"symbol"`
}

func queryBancorInfo(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var param QueryBancorInfoParam
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &param); err != nil {
		return nil, sdk.NewError(types.CodeSpaceBancorlite, types.CodeUnMarshalFailed, "failed to parse param")
	}
	bi := keeper.Load(ctx, param.Symbol)
	var biD BancorInfoDisplay
	if bi != nil {
		biD = NewBancorInfoDisplay(bi)
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, biD)
	if err != nil {
		return nil, types.ErrMarshalFailed()
	}
	return bz, nil
}

func queryBancorList(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	infos := k.GetAllBancorInfos(ctx)
	infoList := make([]BancorInfoDisplay, len(infos))

	for i, info := range infos {
		infoList[i] = NewBancorInfoDisplay(info)
	}
	bz, err := codec.MarshalJSONIndent(k.bik.codec, infoList)
	if err != nil {
		return nil, types.ErrMarshalFailed()
	}
	return bz, nil
}

func queryParameters(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}
