package keepers

import (
	"strconv"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	OrderKey          = []byte{0x01}
	MarketKey         = []byte{0x02}
	BestOrderPriceKey = []byte{0x03}
	PoolLiquidityKey  = []byte{0x04}
)

var (
	BUY      = []byte{0x01}
	SELL     = []byte{0x02}
	NonSwap  = []byte{0x03}
	OpenSwap = []byte{0x04}
)

func getLiquidityKey(marketSymbol string, address sdk.AccAddress) []byte {
	return append(append(PoolLiquidityKey, marketSymbol...), address.Bytes()...)
}

// getMarketKey key = prefix | symbol | swapFlag
// value = vault
func getMarketKey(symbol string, isOpenSwap bool) []byte {
	swapByte := OpenSwap
	if !isOpenSwap {
		swapByte = NonSwap
	}
	return append(append(MarketKey, []byte(symbol)...), swapByte...)
}

// getOrderKey key = prefix | swapFlag | side | symbol | orderID
// value =  order info
func getOrderKey(order *types.Order, isOpenSwap bool) []byte {
	swapByte := OpenSwap
	if !isOpenSwap {
		swapByte = NonSwap
	}
	side := BUY
	if !order.IsBuy {
		side = SELL
	}
	orderID := strconv.Itoa(int(order.OrderID))
	return append(append(append(append(OrderKey, swapByte...),
		side...), order.MarketSymbol...), orderID...)
}

// getBestOrderPriceKey key = prefix | isOpenSwap | side | symbol
// value = orderID
func getBestOrderPriceKey(isOpenSwap, isBuy bool, symbol string) []byte {
	swapByte := OpenSwap
	if !isOpenSwap {
		swapByte = NonSwap
	}
	side := BUY
	if !isBuy {
		side = SELL
	}
	return append(append(append(BestOrderPriceKey, swapByte...), side...), symbol...)
}

func getPairKey(pairSymbol string) []byte {
	return []byte(pairSymbol)
}
