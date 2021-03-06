package autoswap

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
)

const (
	StoreKey      = types.StoreKey
	ModuleName    = types.ModuleName
	PoolModuleAcc = types.PoolModuleAcc
)

var (
	NewKeeper     = keepers.NewKeeper
	DefaultParams = types.DefaultParams
)

type (
	Keeper             = keepers.Keeper
	PoolInfo           = keepers.PoolInfo
	MsgAddLiquidity    = types.MsgAddLiquidity
	MsgRemoveLiquidity = types.MsgRemoveLiquidity
	MsgCreateOrder     = types.MsgAutoSwapCreateOrder
	MsgCancelOrder     = types.MsgAutoSwapCancelOrder
)
