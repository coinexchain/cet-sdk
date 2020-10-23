package types

type DealInfo struct {
	HasDealInOrderBook bool
	RemainAmount       uint64
	AmountInToPool     uint64
	DealMoneyInBook    uint64
	DealStockInBook    uint64
}
