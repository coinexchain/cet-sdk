package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeSpaceAutoSwap sdk.CodespaceType = "autoswap"

	// codes
	CodeInvalidAmount         = 1201
	CodeInvalidPricePrecision = 1202
	CodeInvalidPrice          = 1203
	CodeInvalidOrderSender    = 1204
	CodeUnKnownError          = 1205
	CodeInvalidMarket         = 1206
	CodeInvalidOrderID        = 1207
	CodeMarshalFailed         = 1208
	CodeUnMarshalFailed       = 1209
	CodeInvalidPrevKey        = 1210
	CodeInvalidOrderNews      = 1211
)

func ErrInvalidPrice(price string) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPrice, fmt.Sprintf("Invalid order price: %s", price))
}

func ErrInvalidPricePrecision(pricePrecision int) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPricePrecision, fmt.Sprintf("Invalid "+
		"order price precision: %d, expected: []", pricePrecision))
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
