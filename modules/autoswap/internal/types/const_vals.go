package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	MaxPricePrecision = 10
	MaxAmount, _      = sdk.NewIntFromString("5192296858534827628530496329220096") // 1 << 112
)

const (
	BID = 1
	ASK = 2
)
