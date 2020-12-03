package types

import (
	"github.com/coinexchain/cet-sdk/modules/market"
)

const (
	// Query
	// ModuleName is the name of the module
	//ModuleName = "autoswap"
	ModuleName = market.ModuleName

	// StoreKey is string representation of the store key for autoswap
	StoreKey = "autoswap"

	// RouterKey is the message route for autoswap
	RouterKey = ModuleName

	// QuerierRoute is the querier route for autoswap
	QuerierRoute = ModuleName

	DefaultParamspace = ModuleName

	// Kafka topic name
	Topic = ModuleName

	// Pool's module account
	PoolModuleAcc = "autoswap-pool"
)
