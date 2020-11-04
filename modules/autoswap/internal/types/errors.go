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
	CodeInvalidPrevKey         = 1210
	CodeInvalidOrderNews       = 1211
	CodeInvalidToken           = 1212
	CodeInvalidPairFlag        = 1213
	CodePairAlreadyExist       = 1214
	CodePairIsNotExist         = 1215
	CodeInvalidLiquidityAmount = 1216
	CodeAmountOutIsSmall       = 1217
	CodeInvalidSwap            = 1218
)

func ErrInvalidPrice(price string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPrice, fmt.Sprintf("Invalid order price: %s", price))
}

func ErrInvalidAmount(amount sdk.Int) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidAmount, fmt.Sprintf("Invalid order "+
		"amount: %s, expected: [0: %s]", amount.String(), MaxAmount.String()))
}

func ErrInvalidOutputAmount(amount sdk.Int) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidAmount, fmt.Sprintf("Invalid order "+
		"MinOutputAmount: %s, expected: [0: +∞]", amount.String()))
}

func ErrInvalidSender(sender sdk.AccAddress) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderSender, fmt.Sprintf("Invalid order "+
		"sender: %s", sender.String()))
}

func ErrInvalidMarket(market string, isOpenSwap bool, isOpenOrderBook bool) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidMarket, fmt.Sprintf(
		"Invalid market: %s, isOpenSwap: %v, isOpenOrderBook: %v", market, isOpenSwap, isOpenOrderBook))
}

func ErrInvalidSwap(pairs []MarketInfo) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidSwap, fmt.Sprintf("invalid swap: %v", pairs))
}

func ErrInvalidPrevKey(prevKey [3]int64) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPrevKey, fmt.Sprintf(""+
		"prevKey: [%d, %d, %d]", prevKey[0], prevKey[1], prevKey[2]))
}

func ErrInvalidOrderID(orderID int64) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderID, "Not found valid orderID in create_limit_order msg or "+
		"invalid orderID in delete_order msg: %d", orderID)
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

func ErrInvalidPairFlag(reason string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPairFlag, fmt.Sprintf("reason:%s", reason))
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
