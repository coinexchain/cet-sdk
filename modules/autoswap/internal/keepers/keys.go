package keepers

import "github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"

func getOrderKey(order types.Order) []byte {
	return nil
}

func getPairKey(pairSymbol string) []byte {
	return []byte(pairSymbol)
}
