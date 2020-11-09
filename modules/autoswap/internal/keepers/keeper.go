package keepers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
)

type Keeper struct {
	storeKey      sdk.StoreKey
	paramSubspace params.Subspace
	sk            types.SupplyKeeper
	FactoryInterface
	IPairKeeper
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSubspace params.Subspace,
	bk types.ExpectedBankKeeper, sk types.SupplyKeeper) Keeper {

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
		storeKey:         storeKey,
		paramSubspace:    paramSubspace.WithKeyTable(types.ParamKeyTable()),
		sk:               sk,
		FactoryInterface: factoryK,
	}
	k.IPairKeeper = NewPairKeeper(poolK, sk, bk, cdc, storeKey, k.GetTakerFee, k.GetMakerFee, k.GetDealWithPoolFee)
	return k
}

func (keeper *Keeper) SetParams(ctx sdk.Context, params types.Params) {
	keeper.paramSubspace.SetParamSet(ctx, &params)
}

func (keeper *Keeper) GetParams(ctx sdk.Context) (param types.Params) {
	keeper.paramSubspace.GetParamSet(ctx, &param)
	return
}
func (keeper *Keeper) GetTakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(keeper.GetParams(ctx).TakerFeeRateRate).QuoInt64(types.DefaultFeePrecision)
}

func (keeper *Keeper) GetMakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(keeper.GetParams(ctx).MakerFeeRateRate).QuoInt64(types.DefaultFeePrecision)
}
func (keeper *Keeper) GetDealWithPoolFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(keeper.GetParams(ctx).DealWithPoolFeeRate).QuoInt64(types.DefaultFeePrecision)
}
func (keeper *Keeper) GetFeeToValidator(ctx sdk.Context) sdk.Dec {
	param := keeper.GetParams(ctx)
	return sdk.NewDec(param.FeeToValidator).QuoInt64(param.FeeToValidator + param.FeeToPool)
}
func (keeper Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {
	return keeper.sk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}
func (keeper Keeper) SendCoinsFromUserToPool(ctx sdk.Context, senderAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return keeper.sk.SendCoinsFromAccountToModule(ctx, senderAddr, types.PoolModuleAcc, amt)
}
