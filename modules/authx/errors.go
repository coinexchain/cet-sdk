package authx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeSpaceAuthX sdk.CodespaceType = "authx"

	CodeInvalidMinGasPrice sdk.CodeType = 201
)

func ErrInvalidMinGasPrice(msg string) sdk.Error {
	return sdk.NewError(CodeSpaceAuthX, CodeInvalidMinGasPrice, msg)
}
