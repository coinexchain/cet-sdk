package keepers

import (
	"encoding/binary"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cet-sdk/modules/market"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	OrderBookKey        = []byte{0x01}
	MarketKey           = []byte{0x02}
	MarketEndKey        = []byte{0x03}
	PoolLiquidityKey    = []byte{0x05}
	PoolLiquidityEndKey = []byte{0x06}
	BidOrderKey         = []byte{0x07}
	AskOrderKey         = []byte{0x08}
)

var (
	BIDKEY = []byte{0x01}
	ASKKEY = []byte{0x02}
)

func getLiquidityKey(marketSymbol string, address sdk.AccAddress) []byte {
	return append(append(PoolLiquidityKey, marketSymbol...), address.Bytes()...)
}

// getPairKey key = prefix | Symbol
// value = PoolInfo
func getPairKey(symbol string) []byte {
	return append(MarketKey, []byte(symbol)...)
}

// getBidOrderKey key = bidPrefix | tradingPair | side | 0x0 | price | orderIndexInOneBlock | orderID
// value = nil
func getBidOrderKey(order *types.Order) []byte {
	index := make([]byte, 4)
	binary.BigEndian.PutUint32(index, uint32(order.OrderIndexInOneBlock))
	side := BIDKEY
	if !order.IsBuy {
		side = ASKKEY
	}
	return append(append(append(append(append(append(append(BidOrderKey, order.TradingPair...), side...), 0x0),
		market.DecToBigEndianBytes(order.Price)...)), index...), order.GetOrderID()...)
}

// getBidQueueBegin bidKey = bidPrefix | tradingPair | side | 0x0 | price | orderIndexInOneBlock | orderID
// so beginKey = bidPrefix | tradingPair | side | 0x0
func getBidQueueBegin(tradingPair string) []byte {
	return append(append(append(BidOrderKey, tradingPair...), BIDKEY...), []byte{0x0}...)
}

// getBidQueueEnd bidKey = bidPrefix | tradingPair | side | 0x0 | price | orderIndexInOneBlock | orderID
// so endKey = bidPrefix | tradingPair | side | 0x1
func getBidQueueEnd(tradingPair string) []byte {
	return append(append(append(BidOrderKey, tradingPair...), BIDKEY...), []byte{0x1}...)
}

// getAskOrderKey key = askPrefix | tradingPair | side | 0x0 | price | orderIndexInOneBlock | orderID
// value = nil
func getAskOrderKey(order *types.Order) []byte {
	index := make([]byte, 4)
	binary.BigEndian.PutUint32(index, uint32(order.OrderIndexInOneBlock))
	side := BIDKEY
	if !order.IsBuy {
		side = ASKKEY
	}
	return append(append(append(append(append(append(append(AskOrderKey, order.TradingPair...), side...), 0x0),
		market.DecToBigEndianBytes(order.Price)...)), index...), order.GetOrderID()...)
}

// getAskQueueBegin askKey = askPrefix | tradingPair | side | 0x0 | price | orderIndexInOneBlock | orderID
// so beginKey = askPrefix | tradingPair | side | 0x0
func getAskQueueBegin(tradingPair string) []byte {
	return append(append(append(AskOrderKey, tradingPair...), ASKKEY...), []byte{0x0}...)
}

// getAskQueueEnd askKey = askPrefix | tradingPair | side | 0x0 | price | orderIndexInOneBlock | orderID
// so endKey = askPrefix | tradingPair | side | 0x0
func getAskQueueEnd(tradingPair string) []byte {
	return append(append(append(AskOrderKey, tradingPair...), ASKKEY...), []byte{0x1}...)
}

// getPricePos prefix | tradingPair | side | 0x0
// return: beginPos, endPos.
func getPricePos(tradingPair string) []int {
	begin := 3 + len(tradingPair)
	end := begin + market.DecByteCount
	return []int{begin, end}
}

// getOrderPos the preContent = askPrefix | tradingPair | side | 0x0 | price | orderIndexInOneBlock
func getOrderIDPos(tradingPair string) int {
	return 3 + len(tradingPair) + market.DecByteCount + 4
}

// getOrderKey key = orderBookPrefix | orderID
// value = order info
func getOrderBookKey(orderID string) []byte {
	return append(OrderBookKey, orderID...)
}
