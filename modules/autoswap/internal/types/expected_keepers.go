package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

type ExpectedBankKeeper interface {
	SendCoins(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) sdk.Error // to tranfer coins
	FreezeCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error                   // freeze some coins when creating orders
	UnFreezeCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error                 // unfreeze coins and then orders can be executed
}

type ExpectedAccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.Account
	SetAccount(ctx sdk.Context, acc auth.Account)
}

// SupplyKeeper defines the expected supply keeper (noalias)
type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type ExpectedAuthXKeeper interface {
	GetRefereeAddr(ctx sdk.Context, accAddr sdk.AccAddress) sdk.AccAddress
	GetRebateRatio(ctx sdk.Context) int64
	GetRebateRatioBase(ctx sdk.Context) int64
}
