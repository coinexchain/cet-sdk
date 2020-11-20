package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DealInfo struct {
	// The user will pay amount of token in the order.
	// eg: buy order the amount is money amount, sell order the amount is stock amount.
	RemainAmount      sdk.Int
	AmountInToPool    sdk.Int
	DealMoneyInBook   sdk.Int
	DealStockInBook   sdk.Int
	FeeToMoneyReserve sdk.Int
	FeeToStockReserve sdk.Int
}

func (d DealInfo) String() string {
	return fmt.Sprintf("RemainAmount: %s, AmountInToPool: %s, DealMoneyInBook: %s,"+
		" DealStockInBook: %s, FeeToMoneyReserve: %s, FeeToStockReserve: %s", d.RemainAmount,
		d.AmountInToPool, d.DealMoneyInBook, d.DealStockInBook, d.FeeToMoneyReserve, d.FeeToStockReserve)
}
