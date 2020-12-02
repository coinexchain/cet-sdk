package autoswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cet-sdk/modules/market"
	dex "github.com/coinexchain/cet-sdk/types"
)

// convert msg and redirect to NewHandler()
func NewHandler(k keepers.Keeper) sdk.Handler {
	h1 := NewInternalHandler(k)
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		msg2, ok := convertMarketMsg(msg)
		if !ok {
			return dex.ErrUnknownRequest(types.ModuleName, msg)
		}

		return h1(ctx, msg2)
	}
}

func convertMarketMsg(msg sdk.Msg) (msg2 sdk.Msg, ok bool) {
	ok = true
	switch msg := msg.(type) {
	// market messages
	case market.MsgCreateTradingPair:
		msg2 = convertMsgCreateTP(msg)
	case market.MsgCancelTradingPair:
		msg2 = convertMsgCancelTP(msg)
	//case market.MsgModifyPricePrecision:
	//	panic("TODO")
	case market.MsgCreateOrder:
		msg2 = convertMsgCreateOrder(msg)
	case market.MsgCancelOrder:
		msg2 = convertMsgCancelOrder(msg)
	// new messages
	case types.MsgAddLiquidity, types.MsgRemoveLiquidity:
		msg2 = msg
	default:
		ok = false
	}
	return
}

func convertMsgCreateTP(msg market.MsgCreateTradingPair) types.MsgCreateTradingPair {
	return types.MsgCreateTradingPair{
		Creator:        msg.Creator,
		Stock:          msg.Stock,
		Money:          msg.Money,
		PricePrecision: msg.PricePrecision,
	}
}

func convertMsgCancelTP(msg market.MsgCancelTradingPair) types.MsgCancelTradingPair {
	return types.MsgCancelTradingPair{
		Sender:        msg.Sender,
		TradingPair:   msg.TradingPair,
		EffectiveTime: msg.EffectiveTime,
	}
}

func convertMsgCreateOrder(msg market.MsgCreateOrder) types.MsgCreateOrder {
	return types.MsgCreateOrder{
		Sender:         msg.Sender,
		Identify:       msg.Identify,
		TradingPair:    msg.TradingPair,
		PricePrecision: msg.PricePrecision,
		Price:          msg.Price,
		Quantity:       msg.Quantity,
		Side:           msg.Side,
	}
}
func convertMsgCancelOrder(msg market.MsgCancelOrder) types.MsgCancelOrder {
	return types.MsgCancelOrder{
		Sender:  msg.Sender,
		OrderID: msg.OrderID,
	}
}
