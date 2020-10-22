package keepers

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	PoolLiquidityKey  = []byte{0x01}
)

func getLiquidityKey(marketSymbol string, address sdk.AccAddress) []byte {
	return append(append(PoolLiquidityKey,marketSymbol...), address.Bytes()...)
}

func getOrderKey(order types.Order) []byte {
	return nil
}

func getPairKey(pairSymbol string) []byte {
	return []byte(pairSymbol)
}
