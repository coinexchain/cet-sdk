package alias

import (
	"github.com/coinexchain/cet-sdk/modules/alias/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/alias/internal/types"
)

const (
	StoreKey   = types.StoreKey
	ModuleName = types.ModuleName
)

var (
	ModuleCdc     = types.ModuleCdc
	NewBaseKeeper = keepers.NewKeeper
	DefaultParams = types.DefaultParams
)

type (
	Keeper         = keepers.Keeper
	MsgAliasUpdate = types.MsgAliasUpdate
)
