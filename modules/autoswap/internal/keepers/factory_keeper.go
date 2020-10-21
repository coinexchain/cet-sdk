package keepers

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type FactoryInterface interface {
	CreatePair(ctx sdk.Context, msg types.MsgCreatePair) bool
	QueryPair(ctx sdk.Context, marketSymbol string) *types.MsgCreatePair
}

type Factory struct {
	storeKey sdk.StoreKey
}

func (f Factory) CreatePair(ctx sdk.Context, msg types.MsgCreatePair) bool {
	panic("implement me")
}

func (f Factory) QueryPair(ctx sdk.Context, marketSymbol string) *types.MsgCreatePair {
	panic("implement me")
}
