package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//expected fee collection keeper
type FeeCollectionKeeper interface {
	AddCollectedFees(sdk.Context, sdk.Coins) sdk.Coins
}

// expected bank keeper
type BankKeeper interface {
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
	HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// SupplyKeeper defines the expected supply keeper (noalias)
type SupplyKeeper interface {
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
}

type AssetKeeper interface {
	MintTokenByModule(ctx sdk.Context, symbol string, amount sdk.Int, moduleAddr string) sdk.Error
	BurnTokenByModule(ctx sdk.Context, symbol string, amount sdk.Int, moduleAddr string) sdk.Error
	UpdateCETMintable(ctx sdk.Context) sdk.Error
}
