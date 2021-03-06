package keepers

import (
	"github.com/coinexchain/cet-sdk/msgqueue"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
)

type Keeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey
	sk       types.SupplyKeeper
	FactoryInterface
	IPairKeeper
	types.ExpectedAssetKeeper
	msgProducer msgqueue.MsgSender
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey,
	paramSubspace params.Subspace, bk types.ExpectedBankKeeper, accK types.ExpectedAccountKeeper,
	accxK types.ExpectedAuthXKeeper, sk types.SupplyKeeper, assetK types.ExpectedAssetKeeper) *Keeper {

	poolK := PoolKeeper{
		key:          storeKey,
		codec:        cdc,
		SupplyKeeper: sk,
	}

	factoryK := FactoryKeeper{
		storeKey:   storeKey,
		poolKeeper: poolK,
	}

	k := Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		sk:                  sk,
		FactoryInterface:    factoryK,
		ExpectedAssetKeeper: assetK,
	}
	k.IPairKeeper = NewPairKeeper(poolK, sk, bk, accK, accxK, cdc, storeKey, paramSubspace)
	return &k
}

func (keeper Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {
	return keeper.sk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

func (keeper Keeper) SendCoinsFromUserToPool(ctx sdk.Context, senderAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return keeper.sk.SendCoinsFromAccountToModule(ctx, senderAddr, types.PoolModuleAcc, amt)
}

func (keeper Keeper) SendCoinsFromPoolToUser(ctx sdk.Context, receiver sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return keeper.sk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, receiver, amt)
}

func (keeper Keeper) IsSubscribed(topic string) bool {
	if keeper.msgProducer == nil {
		return false
	}
	return keeper.msgProducer.IsSubscribed(topic)
}
