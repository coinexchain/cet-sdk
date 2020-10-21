package keepers

import sdk "github.com/cosmos/cosmos-sdk/types"

type Pair struct {
	OrderBookKeeperInterface
	//swap
	isOpenSwap bool
}

func NewPair(storeKey sdk.StoreKey, marketSymbol string, isOpenSwap bool) Pair {
	return Pair{
		OrderBookKeeperInterface: OrderBookKeeper{
			marketSymbol: marketSymbol,
			storeKey:     storeKey,
		},
		isOpenSwap: isOpenSwap,
	}
}
