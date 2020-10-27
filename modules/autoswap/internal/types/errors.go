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
)

func ErrInvalidPrice(price int64) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPrice, fmt.Sprintf("Invalid order price: %d", price))
}

func ErrInvalidPricePrecision(pricePrecision int) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidPricePrecision, fmt.Sprintf("Invalid "+
		"order price precision: %d, expected: []", pricePrecision))
}

func ErrInvalidAmount(amount sdk.Int) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidAmount, fmt.Sprintf("Invalid order "+
		"amount: %s, expected: [0: %s]", amount.String(), MaxAmount.String()))
}

func ErrInvalidSender(sender sdk.AccAddress) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderSender, fmt.Sprintf("Invalid order "+
		"sender: %s", sender.String()))
}

func ErrInvalidMarket(market string, isOpenSwap bool) sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidMarket, fmt.Sprintf("Invalid market: %s, isOpenSwap: %v", market, isOpenSwap))
}

func ErrInvalidOrderID() sdk.Error {
	return sdk.NewError(CodeSpaceAutoSwap, CodeInvalidOrderID, "Not found valid orderID")
}
