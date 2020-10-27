package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreateLimitOrder{}

type MsgCreateLimitOrder struct {
	OrderBasic
	OrderID        uint64
	Price          uint64
	PricePrecision byte
	PrevKey        [3]int64
}

func (limit *MsgCreateLimitOrder) Route() string {
	panic("implement me")
}

func (limit *MsgCreateLimitOrder) Type() string {
	panic("implement me")
}

func (limit *MsgCreateLimitOrder) ValidateBasic() sdk.Error {
	if limit.Sender.Empty() || limit.Price == 0 || limit.Amount == 0 {
		return sdk.NewError(RouterKey, 1, "MsgCreateMarketOrder invalid")
	}
	return nil
}

func (limit *MsgCreateLimitOrder) GetSignBytes() []byte {
	panic("implement me")
}

func (limit *MsgCreateLimitOrder) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

func (limit *MsgCreateLimitOrder) String() string {
	content := fmt.Sprintf("Sender: %s, Price: %d, PricePrecision: %d,Amount: "+
		"%d, OrderID: %d\n", limit.Sender.String(), limit.Price, limit.PricePrecision, limit.Amount, limit.OrderID)
	return content
}

func (limit *MsgCreateLimitOrder) SetAccAddress(address sdk.AccAddress) {
	limit.Sender = address
}

var _ sdk.Msg = MsgCreateMarketOrder{}

type MsgCreateMarketOrder struct {
	OrderBasic
}

func (mkOr MsgCreateMarketOrder) Route() string {
	panic("implement me")
}

func (mkOr MsgCreateMarketOrder) Type() string {
	panic("implement me")
}

func (mkOr MsgCreateMarketOrder) ValidateBasic() sdk.Error {
	if mkOr.Sender.Empty() || mkOr.Amount == 0 {
		return sdk.NewError(RouterKey, 2, "MsgCreateMarketOrder invalid")
	}
	return nil
}

func (mkOr MsgCreateMarketOrder) GetSignBytes() []byte {
	panic("implement me")
}

func (mkOr MsgCreateMarketOrder) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

func (mkOr MsgCreateMarketOrder) String() string {
	return fmt.Sprintf("Sender: %s, MarketSymbol: %s, Amount: %d, IsBuy: %v\n",
		mkOr.Sender.String(), mkOr.MarketSymbol, mkOr.Amount, mkOr.IsBuy)
}

func (mkOr *MsgCreateMarketOrder) SetAccAddress(address sdk.AccAddress) {
	mkOr.Sender = address
}

var _ sdk.Msg = MsgCreatePair{}

type MsgCreatePair struct {
	Owner      sdk.AccAddress `json:"owner"`
	Stock      string         `json:"stock"`
	Money      string         `json:"money"`
	StockIn    sdk.Int        `json:"stock_in"`
	MoneyIn    sdk.Int        `json:"money_in"`
	IsOpenSwap bool           `json:"is_open_swap"`
	To         sdk.AccAddress `json:"to"`
}

func (m MsgCreatePair) Route() string {
	return RouterKey
}

func (m MsgCreatePair) Type() string {
	return "create_pair"
}

func (m MsgCreatePair) ValidateBasic() sdk.Error {
	if m.Owner.Empty() {
		return sdk.ErrInvalidAddress("missing owner address")
	}
	if len(m.Stock) == 0 || len(m.Money) == 0 {
		//todo:
		return nil
	}
	if m.StockIn.IsZero() && m.MoneyIn.IsPositive() || m.MoneyIn.IsZero() && m.StockIn.IsPositive() {
		return nil
	}
	//if To is nil, Owner => To
	return nil
}

func (m MsgCreatePair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgCreatePair) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

func (m *MsgCreatePair) SetAccAddress(address sdk.AccAddress) {
	m.Owner = address
}
