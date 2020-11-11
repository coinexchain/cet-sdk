package autoswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cet-sdk/modules/market"
	dex "github.com/coinexchain/cet-sdk/types"
)

func NewHandler2(k keepers.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		// market messages
		case market.MsgCreateTradingPair:
			return handleMsgCreateTradingPair(ctx, k, convertMsgCreateTP(msg))
		case market.MsgCancelTradingPair:
			panic("TODO")
		case market.MsgModifyPricePrecision:
			panic("TODO")
		case market.MsgCreateOrder:
			return handleMsgCreateOrder(ctx, k, convertMsgCreateOrder(msg))
		case market.MsgCancelOrder:
			return handleMsgCancelOrder(ctx, k, convertMsgCancelOrder(msg))
		// new messages
		case types.MsgAddLiquidity:
			return handleMsgAddLiquidity(ctx, k, msg)
		case types.MsgRemoveLiquidity:
			return handleMsgRemoveLiquidity(ctx, k, msg)
		default:
			return dex.ErrUnknownRequest(types.ModuleName, msg)
		}
	}
}

func convertMsgCreateTP(msg market.MsgCreateTradingPair) types.MsgCreateTradingPair {
	return types.MsgCreateTradingPair{
		Creator:        msg.Creator,
		Stock:          msg.Stock,
		Money:          msg.Money,
		PricePrecision: msg.PricePrecision,
		OrderPrecision: msg.OrderPrecision,
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
