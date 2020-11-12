package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type DealInfo struct {
	// The user will pay amount of token in the order.
	// eg: buy order the amount is money amount, sell order the amount is stock amount.
	RemainAmount    sdk.Int
	AmountInToPool  sdk.Int
	DealMoneyInBook sdk.Int
	DealStockInBook sdk.Int
}
