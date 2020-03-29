package authx_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/authx"
	"github.com/coinexchain/cet-sdk/modules/authx/internal/types"
	"github.com/coinexchain/cet-sdk/testutil"
)

var (
	sender  = testutil.ToAccAddress("sender")
	referee = testutil.ToAccAddress("referee")
)

func Test_HandleMsg(t *testing.T) {
	input := setupTestInput()
	input.axk.SetParams(input.ctx, authx.DefaultParams())
	senderAcc := input.ak.NewAccountWithAddress(input.ctx, sender)
	refereeAcc := input.ak.NewAccountWithAddress(input.ctx, referee)
	input.ak.SetAccount(input.ctx, senderAcc)
	input.ak.SetAccount(input.ctx, refereeAcc)
	handler := authx.NewHandler(input.axk, input.ak)
	msg := types.NewMsgSetReferee(sender, referee)

	res := handler(input.ctx, msg)
	require.True(t, res.IsOK())
}

func Test_HandleMsg_AccNotExist(t *testing.T) {
	input := setupTestInput()
	input.axk.SetParams(input.ctx, authx.DefaultParams())

	handler := authx.NewHandler(input.axk, input.ak)
	msg := types.NewMsgSetReferee(sender, referee)

	res := handler(input.ctx, msg)
	require.Equal(t, sdk.CodeUnknownAddress, res.Code)
}

func Test_HandleMsg_RefereeMemoRequired(t *testing.T) {

	input := setupTestInput()
	input.axk.SetParams(input.ctx, authx.DefaultParams())
	senderAcc := input.ak.NewAccountWithAddress(input.ctx, sender)
	refereeAcc := input.ak.NewAccountWithAddress(input.ctx, referee)
	input.ak.SetAccount(input.ctx, senderAcc)
	input.ak.SetAccount(input.ctx, refereeAcc)

	refereeAccx := types.AccountX{
		Address:      referee,
		MemoRequired: true,
	}
	input.axk.SetAccountX(input.ctx, refereeAccx)

	handler := authx.NewHandler(input.axk, input.ak)
	msg := types.NewMsgSetReferee(sender, referee)

	res := handler(input.ctx, msg)
	require.Equal(t, types.CodeRefereeMemoRequired, res.Code)
}

func Test_HandleMsg_RefereeChangeTooFast(t *testing.T) {
	input := setupTestInput()
	input.axk.SetParams(input.ctx, authx.DefaultParams())
	senderAcc := input.ak.NewAccountWithAddress(input.ctx, sender)
	refereeAcc := input.ak.NewAccountWithAddress(input.ctx, referee)
	input.ak.SetAccount(input.ctx, senderAcc)
	input.ak.SetAccount(input.ctx, refereeAcc)

	handler := authx.NewHandler(input.axk, input.ak)
	msg := types.NewMsgSetReferee(sender, referee)

	res := handler(input.ctx, msg)
	require.True(t, res.IsOK())

	//update senderAccx
	senderAccx := types.AccountX{
		Address:           sender,
		RefereeChangeTime: input.ctx.BlockTime().UnixNano(),
	}
	input.axk.SetAccountX(input.ctx, senderAccx)

	//referee changes too fast
	res = handler(input.ctx, msg)
	require.Equal(t, types.CodeRefereeChangeTooFast, res.Code)

	//referee can be updated later
	ctx := input.ctx.WithBlockTime(time.Unix(0, input.ctx.BlockTime().UnixNano()+8*24*time.Hour.Nanoseconds()))
	res = handler(ctx, msg)
	require.True(t, res.IsOK())
}
