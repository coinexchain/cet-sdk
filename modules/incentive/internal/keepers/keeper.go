package keepers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"reflect"

	"github.com/coinexchain/cet-sdk/modules/incentive/internal/types"
)

var (
	StateKey = []byte{0x01}
)

type Keeper struct {
	cdc           *codec.Codec
	key           sdk.StoreKey
	paramSubspace params.Subspace
	types.BankKeeper
	types.SupplyKeeper
	feeCollectorName string
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSubspace params.Subspace,
	bk types.BankKeeper, supplyKeeper types.SupplyKeeper, feeCollectorName string) Keeper {

	return Keeper{
		cdc:              cdc,
		key:              key,
		paramSubspace:    paramSubspace.WithKeyTable(types.ParamKeyTable()),
		BankKeeper:       bk,
		SupplyKeeper:     supplyKeeper,
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
