package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
}

func (m MsgCreatePair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgCreatePair) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}
