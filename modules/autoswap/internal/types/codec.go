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
	cdc.RegisterConcrete(MsgAutoSwapCreateTradingPair{}, "autoswap/MsgAutoSwapCreateTradingPair", nil)
	cdc.RegisterConcrete(MsgAddLiquidity{}, "autoswap/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "autoswap/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(MsgAutoSwapCreateOrder{}, "autoswap/MsgAutoSwapCreateOrder", nil)
	cdc.RegisterConcrete(MsgAutoSwapCancelOrder{}, "autoswap/MsgAutoSwapCancelOrder", nil)
	registerMarketMsgCodec(cdc)
}

func registerMarketMsgCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(market.Order{}, "autoswap/Order", nil)
	cdc.RegisterConcrete(market.MarketInfo{}, "autoswap/TradingPair", nil)
	cdc.RegisterConcrete(market.MsgCreateTradingPair{}, "autoswap/MsgCreateTradingPair", nil)
	cdc.RegisterConcrete(market.MsgCreateOrder{}, "autoswap/MsgCreateOrder", nil)
	cdc.RegisterConcrete(market.MsgCancelOrder{}, "autoswap/MsgCancelOrder", nil)
	cdc.RegisterConcrete(market.MsgCancelTradingPair{}, "autoswap/MsgCancelTradingPair", nil)
	cdc.RegisterConcrete(market.MsgModifyPricePrecision{}, "autoswap/MsgModifyPricePrecision", nil)
}
