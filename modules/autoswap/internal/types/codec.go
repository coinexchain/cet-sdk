package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/coinexchain/cet-sdk/modules/market"
)

var (
	ModuleCdc = codec.New()
)

func init() {
	RegisterCodec(ModuleCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	market.RegisterCodec(cdc)
	cdc.RegisterConcrete(MsgCreateTradingPair{}, "autoswap/MsgCreateTradingPair", nil)
	cdc.RegisterConcrete(MsgAddLiquidity{}, "autoswap/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "autoswap/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(MsgCreateOrder{}, "autoswap/MsgCreateOrder", nil)
	cdc.RegisterConcrete(MsgCancelOrder{}, "autoswap/MsgCancelOrder", nil)
}
