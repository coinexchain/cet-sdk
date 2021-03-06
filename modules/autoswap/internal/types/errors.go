package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeSpaceAutoSwap sdk.CodespaceType = "autoswap"

	// codes
	CodeInvalidAmount          = 1201
	CodeNotFundOrder           = 1202
	CodeInvalidPrice           = 1203
	CodeInvalidOrderSender     = 1204
	CodeUnKnownError           = 1205
	CodeInvalidMarket          = 1206
	CodeInvalidOrderID         = 1207
	CodeMarshalFailed          = 1208
	CodeUnMarshalFailed        = 1209
	CodeInvalidOrderAmount     = 1210
	CodeInvalidOrderNews       = 1211
	CodeInvalidToken           = 1212
	CodeInvalidPairSymbol      = 1213
	CodePairAlreadyExist       = 1214
	CodePairIsNotExist         = 1215
	CodeInvalidLiquidityAmount = 1216
	CodeAmountOutIsSmall       = 1217
	CodeInvalidSwap            = 1218
	CodeInvalidEffectTime      = 1219
	CodeInvalidPricePrecision  = 1220
	CodeOrderAlreadyExist      = 1221
	CodeInvalidOrderSide       = 1222
)

func ErrInvalidPrice(price int64) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPrice, fmt.Sprintf("Invalid order price: %d", price))
}

func ErrInvalidPricePrecision(pricePrecision byte, marketPricePrecision byte) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPricePrecision, fmt.Sprintf("Invalid order price precision: %d,"+
		" market max price precision: %d", pricePrecision, marketPricePrecision))
}

func ErrInvalidAmount(amount sdk.Int) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidAmount, fmt.Sprintf("Invalid order "+
		"amount: %s, expected: [0: %s]", amount.String(), MaxAmount.String()))
}

func ErrInvalidOrderSender(sender sdk.AccAddress) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderSender, fmt.Sprintf("Invalid order "+
		"sender: %s", sender.String()))
}

func ErrInvalidSender(sender, expected sdk.AccAddress) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderSender, fmt.Sprintf("Invalid sender "+
		"sender: %s, expected: %s", sender.String(), expected.String()))
}

func ErrInvalidMarket(market string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidMarket, fmt.Sprintf(
		"Invalid market: %s", market))
}

func ErrInvalidOrderAmount(amount sdk.Int) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderAmount, fmt.Sprintf(
		"Invalid order amount: %s", amount.String()))
}

func ErrInvalidOrderID(orderID string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderID, "Not found valid orderID in create_limit_order msg or "+
		"invalid orderID in delete_order msg: %s", orderID)
}

func ErrMarshalFailed() sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeMarshalFailed, "could not marshal result to JSON")
}

func ErrInvalidOrderNews(orderInfo string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderNews, fmt.Sprintf("received order news: %s", orderInfo))
}

func ErrNotFoundOrder(reason string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeNotFundOrder, fmt.Sprintf("reason: %s", reason))
}

func ErrInvalidToken(reason string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidToken, fmt.Sprintf("reason:%s", reason))
}

func ErrInvalidPairSymbol(reason string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPairSymbol, fmt.Sprintf("reason:%s", reason))
}

func ErrPairAlreadyExist() sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodePairAlreadyExist, "pair already exist")
}

func ErrPairIsNotExist() sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodePairIsNotExist, "pari is not exist")
}

func ErrInvalidLiquidityAmount() sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidLiquidityAmount, "invalid liquidity amount")
}

func ErrAmountOutIsSmallerThanExpected(expected, actual sdk.Int) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeAmountOutIsSmall, fmt.Sprintf("amount out is smaller than "+
		"expected; actual:%s, expected: %s", actual.String(), expected.String()))
}

func ErrInvalidEffectiveTime() sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidEffectTime, "invalid cancel trading pair effectTime")
}

func ErrOrderAlreadyExist(orderID string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeOrderAlreadyExist, fmt.Sprintf("the order already exist: %s", orderID))
}

func ErrInvalidOrderSide(side byte) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderSide, fmt.Sprintf("invalid order side: %d, expected BUY: %d, SELL: %d", side, BID, ASK))
}
