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
	NewKeeper = keepers.NewKeeper
)

type (
	Keeper               = keepers.Keeper
	MsgAddLiquidity      = types.MsgAddLiquidity
	MsgRemoveLiquidity   = types.MsgRemoveLiquidity
	MsgCreateMarketOrder = types.MsgCreateMarketOrder
	MsgCreateLimitOrder  = types.MsgCreateLimitOrder
	MsgDeleteOrder       = types.MsgDeleteOrder
)
