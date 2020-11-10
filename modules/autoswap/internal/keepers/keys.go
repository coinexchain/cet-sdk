package keepers

import (
	"strconv"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	OrderKey          = []byte{0x01}
	MarketKey         = []byte{0x02}
	MarketEndKey      = []byte{0x03}
	BestOrderPriceKey = []byte{0x03}
	PoolLiquidityKey  = []byte{0x04}
)

var (
	BUY  = []byte{0x01}
	SELL = []byte{0x02}
)

func getLiquidityKey(marketSymbol string, address sdk.AccAddress) []byte {
	return append(append(PoolLiquidityKey, getPairKey(marketSymbol)...), address.Bytes()...)
}

// getPairKey key = prefix | Symbol
// value = PoolInfo
func getPairKey(symbol string) []byte {
	return append(MarketKey, []byte(symbol)...)
}

// getOrderKey key = prefix | side | Symbol | orderID
// value = pair info
func getOrderKey(order *types.Order) []byte {
	side := BUY
	if !order.IsBuy {
		side = SELL
	}
	orderID := strconv.Itoa(int(order.OrderID))
	return append(append(append(OrderKey, side...), order.MarketSymbol...), orderID...)
}

// getBestOrderPriceKey key = prefix | side | Symbol
// value = orderID
func getBestOrderPriceKey(symbol string, isBuy bool) []byte {
	side := BUY
	if !isBuy {
		side = SELL
	}
	return append(append(BestOrderPriceKey, side...), symbol...)
}
