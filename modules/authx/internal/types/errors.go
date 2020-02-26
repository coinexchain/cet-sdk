package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeSpaceAuthX sdk.CodespaceType = "authx"

	// 201 ～ 299
	CodeInvalidMinGasPriceLimit sdk.CodeType = 201
	CodeGasPriceTooLow          sdk.CodeType = 202
	CodeRefereeChangeTooFast    sdk.CodeType = 203
	CodeRefereeMemoRequired     sdk.CodeType = 204
	CodeRefereeCanNotBeYourself sdk.CodeType = 205
)

func ErrInvalidMinGasPriceLimit(limit sdk.Dec) sdk.Error {
	return sdk.NewError(CodeSpaceAuthX, CodeInvalidMinGasPriceLimit,
		"invalid minimum gas price limit: %s", limit)
}

func ErrGasPriceTooLow(required, actual sdk.Dec) sdk.Error {
	return sdk.NewError(CodeSpaceAuthX, CodeGasPriceTooLow,
		"gas price too low: %s < %s", actual, required)
}
func ErrRefereeChangeTooFast(referee string) sdk.Error {
	return sdk.NewError(CodeSpaceAuthX, CodeRefereeChangeTooFast, "refere %s change too fast", referee)
}
func ErrRefereeMemoRequired(referee string) sdk.Error {
	return sdk.NewError(CodeSpaceAuthX, CodeRefereeMemoRequired, "referee %s must not be memo required", referee)
}
func ErrRefereeCanNotBeYouself(referee string) sdk.Error {
	return sdk.NewError(CodeSpaceAuthX, CodeRefereeCanNotBeYourself, "referee %s can not be yourself", referee)
}
