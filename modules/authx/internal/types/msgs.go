package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = MsgSetReferee{}

type MsgSetReferee struct {
	Sender  sdk.AccAddress `json:"sender"`
	Referee sdk.AccAddress `json:"referee"`
}

func NewMsgSetReferee(sender sdk.AccAddress, referee sdk.AccAddress) MsgSetReferee {
	return MsgSetReferee{Sender: sender, Referee: referee}
}

func (msg *MsgSetReferee) SetAccAddress(addr sdk.AccAddress) {
	msg.Sender = addr
}

// --------------------------------------------------------
// sdk.Msg Implementation

func (msg MsgSetReferee) Route() string { return RouteKey }

func (msg MsgSetReferee) Type() string { return "set_referee_address" }

func (msg MsgSetReferee) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() || msg.Referee.Empty() {
		return sdk.ErrInvalidAddress("missing address")
	}
	if msg.Sender.Equals(msg.Referee) {
		return ErrRefereeCanNotBeYouself(msg.Referee.String())
	}
	return nil
}

func (msg MsgSetReferee) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgSetReferee) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
