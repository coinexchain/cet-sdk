package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type CreateOrderInfo struct {
	OrderID     string  `json:"order_id"`
	Sender      string  `json:"sender"`
	TradingPair string  `json:"trading_pair"`
	Price       sdk.Dec `json:"price"`
	Quantity    int64   `json:"quantity"`
	Side        byte    `json:"side"`
	Height      int64   `json:"height"`
	Freeze      int64   `json:"freeze"`
}

type FillOrderInfo struct {
	OrderID     string  `json:"order_id"`
	TradingPair string  `json:"trading_pair"`
	Height      int64   `json:"height"`
	Side        byte    `json:"side"`
	Price       sdk.Dec `json:"price"`

	// These fields will change when order was filled/canceled.
	LeftStock          int64   `json:"left_stock"`
	Freeze             int64   `json:"freeze"`
	DealStock          int64   `json:"deal_stock"`
	DealMoney          int64   `json:"deal_money"`
	CurrStock          int64   `json:"curr_stock"`
	CurrMoney          int64   `json:"curr_money"`
	FillPrice          sdk.Dec `json:"fill_price"`
	CurrUsedCommission int64   `json:"curr_used_commission"`
}

type MarketDealInfo struct {
	TradingPair     string
	MakerOrderID    string
	TakerOrderID    string
	DealStockAmount int64
	DealHeight      int64
}

type CancelOrderInfo struct {
	OrderID     string  `json:"order_id"`
	TradingPair string  `json:"trading_pair"`
	Height      int64   `json:"height"`
	Side        byte    `json:"side"`
	Price       sdk.Dec `json:"price"`

	// Del infos
	DelReason string `json:"del_reason"`

	// Fields of amount
	RebateAmount      int64  `json:"rebate_amount"`
	RebateRefereeAddr string `json:"rebate_referee_addr"`
	LeftStock         int64  `json:"left_stock"`
	RemainAmount      int64  `json:"remain_amount"`
	DealStock         int64  `json:"deal_stock"`
	DealMoney         int64  `json:"deal_money"`
}
