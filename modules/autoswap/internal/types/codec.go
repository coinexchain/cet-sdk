package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	ModuleCdc = codec.New()
)

func init() {
	RegisterCodec(ModuleCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgAddLiquidity{}, "autoswap/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "autoswap/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(MsgCreateOrder{}, "autoswap/MsgCreateOrder", nil)
	cdc.RegisterConcrete(MsgCancelOrder{}, "autoswap/MsgCancelOrder", nil)
}
