package keepers

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"
)

type Keeper struct {
	storeKey      sdk.StoreKey
	paramSubspace params.Subspace
	sk            types.SupplyKeeper
	FactoryInterface
	//IPoolKeeper
	IPairKeeper
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSubspace params.Subspace,
	ak types.ExpectedAccountKeeper, bk types.ExpectedBankKeeper, sk types.SupplyKeeper) Keeper {

	poolK := PoolKeeper{
		key:          storeKey,
		codec:        cdc,
		SupplyKeeper: sk,
	}

	pairK := PairKeeper{
		IPoolKeeper:        poolK,
		SupplyKeeper:       sk,
		ExpectedBankKeeper: bk,
		codec:              cdc,
		storeKey:           storeKey,
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
		//IPoolKeeper:      poolK,
		//IPairKeeper: pairK,
	}
	pairK.GetMakerFee = func(ctx sdk.Context) sdk.Dec { return k.GetMakerFee(ctx) }
	pairK.GetTakerFee = func(ctx sdk.Context) sdk.Dec { return k.GetTakerFee(ctx) }
	k.IPairKeeper = pairK
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
func (keeper *Keeper) GetFeeToValidator(ctx sdk.Context) sdk.Dec {
	param := keeper.GetParams(ctx)
	return sdk.NewDec(param.FeeToValidator).QuoInt64(param.FeeToValidator + param.FeeToPool)
}
func (keeper Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {
	return keeper.sk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}
func (keeper Keeper) SendCoinsFromPoolAccountToModule(ctx sdk.Context, recipientModule string, amt sdk.Coins) sdk.Error {
	poolAddr := keeper.sk.GetModuleAddress(types.PoolModuleAcc)
	return keeper.sk.SendCoinsFromAccountToModule(ctx, poolAddr, recipientModule, amt)
}
func (keeper *Keeper) AllocateFeeToValidator(ctx sdk.Context, lastK *sdk.Int, info PoolInfo) error {
	if !lastK.IsZero() {
		k := info.moneyAmmReserve.Mul(info.stockAmmReserve).BigInt()
		var rootK, rootLastK big.Int
		rootK.Sqrt(k)
		rootLastK.Sqrt(lastK.BigInt())

		if rootK.Cmp(&rootLastK) == 1 {
			var subKBigInt big.Int
			subK := sdk.NewIntFromBigInt(subKBigInt.Sub(&rootK, &rootLastK))
			param := keeper.GetParams(ctx)

			numerator := subK.MulRaw(param.FeeToValidator)
			denominator := sdk.NewIntFromBigInt(&rootK).MulRaw(param.FeeToPool).
				Add(sdk.NewIntFromBigInt(&rootLastK).MulRaw(param.FeeToValidator))
			moneyToValidator := info.moneyAmmReserve.Mul(numerator).Quo(denominator)
			stockToValidator := info.stockAmmReserve.Mul(numerator).Quo(denominator)
			stock, money := dex.SplitSymbol(info.symbol)

			if moneyToValidator.IsPositive() {
				if err := keeper.SendCoinsFromPoolAccountToModule(ctx, auth.FeeCollectorName,
					sdk.NewCoins(sdk.NewCoin(money, moneyToValidator))); err != nil {
					return err
				}
			}
			if stockToValidator.IsPositive() {
				if err := keeper.SendCoinsFromPoolAccountToModule(ctx, auth.FeeCollectorName,
					sdk.NewCoins(sdk.NewCoin(stock, stockToValidator))); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
