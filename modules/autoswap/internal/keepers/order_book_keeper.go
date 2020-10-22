package keepers

import (
	"strings"

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
	marketSymbol string // stock/money
}

func (obk OrderBookKeeper) Stock() string {
	return strings.Split(obk.marketSymbol, "/")[0]
}

func (obk OrderBookKeeper) Money() string {
	return strings.Split(obk.marketSymbol, "/")[1]
}

func (obk OrderBookKeeper) AddMarketOrder(ctx sdk.Context, order *types.Order) bool {

	// 1. 计算订单ID
	// 2. 计算插入位置
	// 3. 计算金额
	// 4. 计算是否可以直接插入订单，如果可以插入，表示该订单当前无法成交
	// 5. 进行订单成交
	// 6. 将成交剩余的订单插入进订单簿

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
