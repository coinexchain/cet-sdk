package keepers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/coinexchain/cet-sdk/modules/bancorlite/internal/types"
	"github.com/coinexchain/cet-sdk/msgqueue"
	dex "github.com/coinexchain/cet-sdk/types"
)

var (
	BancorInfoKey    = []byte{0x10}
	BancorInfoKeyEnd = []byte{0x11}
)

type BancorInfoKeeper struct {
	biKey         sdk.StoreKey
	codec         *codec.Codec
	paramSubspace params.Subspace
}

func NewBancorInfoKeeper(key sdk.StoreKey, cdc *codec.Codec, paramSubspace params.Subspace) *BancorInfoKeeper {
	return &BancorInfoKeeper{
		biKey:         key,
		codec:         cdc,
		paramSubspace: paramSubspace.WithKeyTable(types.ParamKeyTable()),
	}
}

func (keeper *BancorInfoKeeper) SetParams(ctx sdk.Context, params types.Params) {
	keeper.paramSubspace.SetParamSet(ctx, &params)
}

func (keeper *BancorInfoKeeper) GetParams(ctx sdk.Context) (param types.Params) {
	keeper.paramSubspace.GetParamSet(ctx, &param)
	return
}

func (keeper *BancorInfoKeeper) Save(ctx sdk.Context, bi *BancorInfo) {
	store := ctx.KVStore(keeper.biKey)
	value := keeper.codec.MustMarshalBinaryBare(bi)
	key := append(BancorInfoKey, []byte(bi.GetSymbol())...)
	store.Set(key, value)
}

func (keeper *BancorInfoKeeper) Remove(ctx sdk.Context, bi *BancorInfo) {
	store := ctx.KVStore(keeper.biKey)
	key := append(BancorInfoKey, []byte(bi.GetSymbol())...)
	value := store.Get(key)
	if value != nil {
		store.Delete(key)
	}
}

//key: stock/money pair
func (keeper *BancorInfoKeeper) Load(ctx sdk.Context, symbol string) *BancorInfo {
	store := ctx.KVStore(keeper.biKey)
	key := append(BancorInfoKey, []byte(symbol)...)
	biBytes := store.Get(key)
	if biBytes == nil {
		return nil
	}
	bi := &BancorInfo{}
	keeper.codec.MustUnmarshalBinaryBare(biBytes, bi)
	return bi
}

func (keeper *BancorInfoKeeper) Iterate(ctx sdk.Context, biProc func(bi *BancorInfo)) {
	store := ctx.KVStore(keeper.biKey)
	iter := store.Iterator(BancorInfoKey, BancorInfoKeyEnd)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		bi := &BancorInfo{}
		keeper.codec.MustUnmarshalBinaryBare(iter.Value(), bi)
		biProc(bi)
	}
}

type Keeper struct {
	bik         *BancorInfoKeeper
	bxk         types.ExpectedBankxKeeper
	ask         types.ExpectedAssetStatusKeeper
	mk          types.ExpectedMarketKeeper
	axk         types.ExpectedAuthXKeeper
	msgProducer msgqueue.MsgSender
}

func NewKeeper(bik *BancorInfoKeeper,
	bxk types.ExpectedBankxKeeper,
	ask types.ExpectedAssetStatusKeeper,
	mk types.ExpectedMarketKeeper,
	axk types.ExpectedAuthXKeeper,
	mq msgqueue.MsgSender) Keeper {
	return Keeper{
		bik:         bik,
		bxk:         bxk,
		ask:         ask,
		mk:          mk,
		axk:         axk,
		msgProducer: mq,
	}
}

func (keeper Keeper) IsBancorExist(ctx sdk.Context, stock string) bool {
	store := ctx.KVStore(keeper.bik.biKey)
	key := append(BancorInfoKey, []byte(stock+dex.SymbolSeparator)...)
	iter := store.Iterator(key, append(key, 0xff))
	defer iter.Close()
	iter.Domain()
	for iter.Valid() {
		return true
	}
	return false
}

func (keeper *Keeper) GetAllBancorInfos(ctx sdk.Context) (list []*BancorInfo) {
	keeper.Iterate(ctx, func(bi *BancorInfo) {
		list = append(list, bi)
	})
	return
}

func (keeper *Keeper) SetParams(ctx sdk.Context, params types.Params) {
	keeper.bik.SetParams(ctx, params)
}

func (keeper *Keeper) GetParams(ctx sdk.Context) (param types.Params) {
	keeper.bik.paramSubspace.GetParamSet(ctx, &param)
	return
}

func (keeper *Keeper) Save(ctx sdk.Context, bi *BancorInfo) {
	keeper.bik.Save(ctx, bi)
}

func (keeper *Keeper) Remove(ctx sdk.Context, bi *BancorInfo) {
	keeper.bik.Remove(ctx, bi)
}

func (keeper *Keeper) Load(ctx sdk.Context, symbol string) *BancorInfo {
	return keeper.bik.Load(ctx, symbol)
}

func (keeper *Keeper) Iterate(ctx sdk.Context, biProc func(bi *BancorInfo)) {
	keeper.bik.Iterate(ctx, biProc)
}

func (keeper *Keeper) SendCoins(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return keeper.bxk.SendCoins(ctx, from, to, amt)
}
func (keeper *Keeper) FreezeCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return keeper.bxk.FreezeCoins(ctx, acc, amt)
}
func (keeper *Keeper) UnFreezeCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return keeper.bxk.UnFreezeCoins(ctx, acc, amt)
}
func (keeper *Keeper) DeductFee(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return keeper.bxk.DeductFee(ctx, acc, amt)
}
func (keeper *Keeper) DeductInt64CetFee(ctx sdk.Context, addr sdk.AccAddress, amt int64) sdk.Error {
	return keeper.bxk.DeductInt64CetFee(ctx, addr, amt)
}

func (keeper *Keeper) IsTokenExists(ctx sdk.Context, denom string) bool {
	return keeper.ask.IsTokenExists(ctx, denom)
}
func (keeper *Keeper) IsTokenIssuer(ctx sdk.Context, denom string, addr sdk.AccAddress) bool {
	return keeper.ask.IsTokenIssuer(ctx, denom, addr)
}
func (keeper *Keeper) IsForbiddenByTokenIssuer(ctx sdk.Context, denom string, addr sdk.AccAddress) bool {
	return keeper.ask.IsForbiddenByTokenIssuer(ctx, denom, addr)
}

func (keeper *Keeper) GetMarketVolume(ctx sdk.Context, stock, money string, stockVolume, moneyVolume sdk.Dec) sdk.Dec {
	return keeper.mk.GetMarketVolume(ctx, stock, money, stockVolume, moneyVolume)
}

func (keeper *Keeper) IsMarketExist(ctx sdk.Context, symbol string) bool {
	return keeper.mk.IsMarketExist(ctx, symbol)
}
func (keeper *Keeper) GetMarketFeeMin(ctx sdk.Context) int64 {
	return keeper.mk.GetMarketFeeMin(ctx)
}

func (keeper *Keeper) GetRefereeAddr(ctx sdk.Context, accAddr sdk.AccAddress) sdk.AccAddress {
	acc := keeper.axk.GetRefereeAddr(ctx, accAddr)
	if len(acc) == 0 {
		return nil
	}
	return acc
}

func (keeper *Keeper) GetRebateRatio(ctx sdk.Context) int64 {
	return keeper.axk.GetRebateRatio(ctx)
}

func (keeper *Keeper) GetRebateRatioBase(ctx sdk.Context) int64 {
	return keeper.axk.GetRebateRatioBase(ctx)
}

func (keeper *Keeper) GetRebate(ctx sdk.Context, address sdk.AccAddress, total sdk.Int) (acc sdk.AccAddress, rebate, balance sdk.Int, exist bool) {
	acc = keeper.axk.GetRefereeAddr(ctx, address)
	if len(acc) == 0 {
		return acc, sdk.ZeroInt(), sdk.ZeroInt(), false
	}
	ratio := keeper.GetRebateRatio(ctx)
	base := keeper.GetRebateRatioBase(ctx)
	rebate = total.MulRaw(ratio).QuoRaw(base)
	balance = total.Sub(rebate)
	exist = true
	return
}

func (keeper *Keeper) IsSubscribed(topic string) bool {
	return keeper.msgProducer.IsSubscribed(topic)
}
