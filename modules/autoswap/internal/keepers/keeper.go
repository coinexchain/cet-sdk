package keepers

import sdk "github.com/cosmos/cosmos-sdk/types"

type Keeper struct {
	storeKey sdk.StoreKey
	factory  FactoryInterface
	pairs    map[string]Pair // symbol
}
