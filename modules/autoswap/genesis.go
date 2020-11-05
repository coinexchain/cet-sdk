package autoswap

import (
	"errors"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisState struct {
	Params    types.Params        `json:"params"`
	Orders    []*types.Order      `json:"orders"`
	PoolInfos []*keepers.PoolInfo `json:"pool_infos"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params types.Params, orders []*types.Order, infos []*keepers.PoolInfo) GenesisState {
	return GenesisState{
		Params:    params,
		Orders:    orders,
		PoolInfos: infos,
	}
}

func DefaultGenesisState() GenesisState {
	return NewGenesisState(types.DefaultParams(), []*types.Order{}, []*keepers.PoolInfo{})
}

func InitGenesis(ctx sdk.Context, keeper keepers.Keeper, data GenesisState) {
	keeper.SetParams(ctx, data.Params)

	for _ = range data.Orders {
		//todo: setOrders
	}

	for _ = range data.PoolInfos {
		//todo: setPoolInfo
	}
}

func ExportGenesis(ctx sdk.Context, k keepers.Keeper) GenesisState {
	return GenesisState{}
}

func (data GenesisState) Validate() error {
	if err := data.Params.ValidateGenesis(); err != nil {
		return err
	}
	//todo: check duplicate order with id
	infos := make(map[string]struct{})
	for _, info := range data.PoolInfos {
		symbol := info.Symbol
		if _, exists := infos[symbol]; exists {
			return errors.New("duplicate pool found during autoswap genesis validate")
		}
		infos[symbol] = struct{}{}
	}
	return nil
}
