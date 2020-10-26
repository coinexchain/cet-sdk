package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type DealInfo struct {
	HasDealInOrderBook bool
	RemainAmount       sdk.Int
	AmountInToPool     sdk.Int
	DealMoneyInBook    sdk.Int
	DealStockInBook    sdk.Int
}
