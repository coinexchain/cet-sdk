package types

const (
	// StoreKey is string representation of the store key for authx
	StoreKey = "accx"
	// QuerierRoute is the querier route for accx
	QuerierRoute = StoreKey

	RouteKey = ModuleName
)

// query endpoints supported by the auth Querier
const (
	QueryParameters = "parameters"
	QueryAccountMix = "accountMix"
)
