package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/testutil"
)

var (
	sender   = testutil.ToAccAddress("sender")
	referee  = testutil.ToAccAddress("referee")
	noneAddr = sdk.AccAddress{}
)

func TestMsgSetReferee_ValidateBasic(t *testing.T) {

	tests := []struct {
		name string
		msg  MsgSetReferee
		want error
	}{
		{
			name: "basic_test",
			msg:  NewMsgSetReferee(sender, referee),
			want: nil,
		},
		{
			name: "sender_nil",
			msg:  NewMsgSetReferee(noneAddr, referee),
			want: sdk.ErrInvalidAddress("missing address"),
		},
		{
			name: "referee_nil",
			msg:  NewMsgSetReferee(sender, noneAddr),
			want: sdk.ErrInvalidAddress("missing address"),
		},
		{
			name: "referee the same to sender",
			msg:  NewMsgSetReferee(sender, sender),
			want: ErrRefereeCanNotBeYouself(sender.String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.ValidateBasic(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MsgIssueToken.ValidateBasic() = %v, want %v", got, tt.want)
			}
		})
	}

}
func TestMsgSetReferee_GetSigners(t *testing.T) {
	msg := NewMsgSetReferee(sender, referee)
	require.Equal(t, msg.GetSigners(), []sdk.AccAddress{sender})
}
func TestMsgSetReferee_Routesg(t *testing.T) {
	msg := NewMsgSetReferee(sender, referee)
	require.Equal(t, msg.Route(), ModuleName)
}
