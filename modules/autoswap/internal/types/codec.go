package types

import (
	"github.com/coinexchain/cet-sdk/modules/market"
	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	ModuleCdc = codec.New()
)

func init() {
	RegisterCodec(ModuleCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgAutoSwapCreateTradingPair{}, "market/MsgAutoSwapCreateTradingPair", nil)
	cdc.RegisterConcrete(MsgAddLiquidity{}, "market/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "market/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(MsgAutoSwapCreateOrder{}, "market/MsgAutoSwapCreateOrder", nil)
	cdc.RegisterConcrete(MsgAutoSwapCancelOrder{}, "market/MsgAutoSwapCancelOrder", nil)
	market.RegisterCodec(cdc)
}
