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
	BUY           = []byte{0x01}
	SELL          = []byte{0x02}
	NonSwap       = []byte{0x03}
	OpenSwap      = []byte{0x04}
	NonOpenBook   = []byte{0x05}
	OpenOrderBook = []byte{0x06}
)

func getLiquidityKey(marketSymbol string, address sdk.AccAddress) []byte {
	return append(append(PoolLiquidityKey, marketSymbol...), address.Bytes()...)
}

// getPairKey key = prefix | Symbol | swapFlag | openOrderBookFlag
// value = PoolInfo
func getPairKey(symbol string, isOpenSwap, isOpenOrderBook bool) []byte {
	swapByte := OpenSwap
	if !isOpenSwap {
		swapByte = NonSwap
	}
	orderBookByte := NonOpenBook
	if isOpenOrderBook {
		orderBookByte = OpenOrderBook
	}
	return append(append(append(MarketKey, []byte(symbol)...), swapByte...), orderBookByte...)
}

// getOrderKey key = prefix | swapFlag | isOpenOrderBook | side | Symbol | orderID
// value = pair info
func getOrderKey(order *types.Order) []byte {
	swapByte := OpenSwap
	if !order.IsOpenSwap {
		swapByte = NonSwap
	}
	orderBookByte := NonOpenBook
	if order.IsOpenOrderBook {
		orderBookByte = OpenOrderBook
	}
	side := BUY
	if !order.IsBuy {
		side = SELL
	}
	orderID := strconv.Itoa(int(order.OrderID))
	return append(append(append(append(append(OrderKey, swapByte...),
		orderBookByte...), side...), order.MarketSymbol...), orderID...)
}

// getBestOrderPriceKey key = prefix | isOpenSwap | isOpenOrderBook | side | Symbol
// value = orderID
func getBestOrderPriceKey(symbol string, isOpenSwap, isOpenOrderBook, isBuy bool) []byte {
	swapByte := OpenSwap
	if !isOpenSwap {
		swapByte = NonSwap
	}
	orderBookByte := NonOpenBook
	if isOpenOrderBook {
		orderBookByte = OpenOrderBook
	}
	side := BUY
	if !isBuy {
		side = SELL
	}
	return append(append(append(append(BestOrderPriceKey,
		swapByte...), orderBookByte...), side...), symbol...)
}
