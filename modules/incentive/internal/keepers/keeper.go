package keepers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/crypto"
	"reflect"

	"github.com/coinexchain/cet-sdk/modules/incentive/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"
)

var (
	StateKey = []byte{0x01}
	PoolAddr = sdk.AccAddress(crypto.AddressHash([]byte("incentive_pool")))
)

type Keeper struct {
	cdc           *codec.Codec
	key           sdk.StoreKey
	paramSubspace params.Subspace
	types.BankKeeper
	types.SupplyKeeper
	types.AssetKeeper
	feeCollectorName string
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSubspace params.Subspace,
	bk types.BankKeeper, supplyKeeper types.SupplyKeeper, assetKeeper types.AssetKeeper, feeCollectorName string) Keeper {

	return Keeper{
		cdc:              cdc,
		key:              key,
		paramSubspace:    paramSubspace.WithKeyTable(types.ParamKeyTable()),
		BankKeeper:       bk,
		SupplyKeeper:     supplyKeeper,
		AssetKeeper:      assetKeeper,
		feeCollectorName: feeCollectorName,
	}
}

func (k Keeper) GetParams(ctx sdk.Context) (param types.Params) {
	k.paramSubspace.GetParamSet(ctx, &param)
	return
}
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

func (k Keeper) SetUpdatedRewards(ctx sdk.Context, newFixedValue int64) {
	v := reflect.Indirect(reflect.ValueOf(newFixedValue)).Interface()
	k.paramSubspace.Set(ctx, types.KeyIncentiveUpdatedRewards, v)
}

func (k Keeper) GetState(ctx sdk.Context) (state types.State) {
	store := ctx.KVStore(k.key)
	bz := store.Get(StateKey)
	if bz == nil {
		panic("cannot load the adjustment height for incentive pool")
	}
	if err := k.cdc.UnmarshalBinaryBare(bz, &state); err != nil {
		panic(err)
	}
	return
}

func (k Keeper) SetState(ctx sdk.Context, state types.State) sdk.Error {
	store := ctx.KVStore(k.key)
	bz, err := k.cdc.MarshalBinaryBare(state)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}
	store.Set(StateKey, bz)
	return nil
}

func (k Keeper) AddNewPlan(ctx sdk.Context, plan types.Plan) sdk.Error {
	if err := types.CheckPlans([]types.Plan{plan}); err != nil {
		return sdk.NewError(types.CodeSpaceIncentive, types.CodeInvalidPlanToAdd, "new plan is invalid")
	}
	param := k.GetParams(ctx)
	param.Plans = append(param.Plans, plan)
	k.SetParams(ctx, param)
	return nil
}
func (k Keeper) ClearIncentiveState(ctx sdk.Context) {
	// clear unused plans & params
	k.SetParams(ctx, types.DefaultParams())

	// burn pool's cet token
	allCoins := k.GetCoins(ctx, PoolAddr)
	for _, coin := range allCoins {
		if coin.Denom == dex.DefaultBondDenom {
			err := k.SendCoinsFromAccountToModule(ctx, PoolAddr, types.ModuleName, sdk.NewCoins(coin))
			if err != nil {
				panic(err)
			}
			err = k.BurnTokenByModule(ctx, coin.Denom, coin.Amount, types.ModuleName)
			if err != nil {
				panic(err)
			}
		}
	}
	// update cet to be mintable
	err := k.UpdateCETMintable(ctx)
	if err != nil {
		panic(err)
	}
}
