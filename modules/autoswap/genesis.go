package autoswap

import (
	"errors"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisState struct {
	Params         types.Params            `json:"params"`
	Orders         []types.Order           `json:"orders"`
	PoolInfos      []keepers.PoolInfo      `json:"pool_infos"`
	LiquidityInfos []keepers.LiquidityInfo `json:"liquidity_infos"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params types.Params, orders []types.Order, infos []keepers.PoolInfo, liquidityInfos []keepers.LiquidityInfo) GenesisState {
	return GenesisState{
		Params:         params,
		Orders:         orders,
		PoolInfos:      infos,
		LiquidityInfos: liquidityInfos,
	}
}

func DefaultGenesisState() GenesisState {
	return NewGenesisState(types.DefaultParams(), []types.Order{}, []keepers.PoolInfo{}, []keepers.LiquidityInfo{})
}

func InitGenesis(ctx sdk.Context, k *keepers.Keeper, data GenesisState) {
	k.SetParams(ctx, data.Params)
	for _ = range data.Orders {
		//todo: setOrders
	}
	for _, info := range data.PoolInfos {
		k.SetPoolInfo(ctx, info.Symbol, &info)
	}
	for _, li := range data.LiquidityInfos {
		k.SetLiquidity(ctx, li.Symbol, li.Owner, li.Liquidity)
	}
}

func ExportGenesis(ctx sdk.Context, k keepers.Keeper) GenesisState {
	infos := k.GetPoolInfos(ctx)
	var g GenesisState
	g.PoolInfos = infos
	g.Params = k.GetParams(ctx)
	g.LiquidityInfos = k.GetAllLiquidityInfos(ctx)
	return g
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
