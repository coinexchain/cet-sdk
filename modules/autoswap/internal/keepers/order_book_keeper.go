package keepers

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderBookKeeperInterface interface {
	AddMarketOrder(ctx sdk.Context, order *types.Order) bool
	AddLimitOrder(ctx sdk.Context, order *types.Order) bool
	HasOrder(ctx sdk.Context, isBuy bool, orderID uint64) bool
	QueryOrderInfo(ctx sdk.Context, isBuy bool, orderID uint64) *types.Order
}

type OrderBookKeeper struct {
	sellOrders map[uint64]*types.Order
	buyOrders  map[uint64]*types.Order

	storeKey     sdk.StoreKey
	marketSymbol string
}

func (obk OrderBookKeeper) AddMarketOrder(ctx sdk.Context, order *types.Order) bool {
	return true
}

func (obk OrderBookKeeper) AddLimitOrder(ctx sdk.Context, order *types.Order) bool {

	return true
}

func (obk OrderBookKeeper) getUnusedOrderID(ctx sdk.Context, isBuy bool, id uint64) uint64 {

	return 0
}

func (obk OrderBookKeeper) HasOrder(ctx sdk.Context, isBuy bool, orderID uint64) bool {
	return true
}

func (obk OrderBookKeeper) QueryOrderInfo(ctx sdk.Context, isBuy bool, orderID uint64) *types.Order {
	return nil
}
