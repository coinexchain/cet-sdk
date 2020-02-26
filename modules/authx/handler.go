package authx

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/authx/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/authx/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"
)

func NewHandler(k keepers.AccountXKeeper, ak ExpectedAccountKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgSetReferee:
			return handleMsgSetReferee(ctx, k, ak, msg)
		default:
			return dex.ErrUnknownRequest(ModuleName, msg)

		}
	}
}

func handleMsgSetReferee(ctx sdk.Context, k keepers.AccountXKeeper, ak ExpectedAccountKeeper, msg types.MsgSetReferee) sdk.Result {
	if err := preCheckAddr(ctx, k, ak, msg); err != nil {
		return err.Result()
	}

	senderAccx := k.GetOrCreateAccountX(ctx, msg.Sender)
	refereeMinChangeInterval := k.GetParams(ctx).RefereeChangeMinInterval

	if err := preCheckTime(ctx.BlockTime().UnixNano(), senderAccx.RefereeChangeTime, refereeMinChangeInterval, msg.Referee); err != nil {
		return err.Result()
	}

	senderAccx.UpdateRefereeAddr(msg.Referee, ctx.BlockTime().UnixNano())
	k.SetAccountX(ctx, senderAccx)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	})
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(types.EventTypeSetReferee,
			sdk.NewAttribute(types.AttributeReferee, msg.Referee.String()),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
func preCheckAddr(ctx sdk.Context, k keepers.AccountXKeeper, ak ExpectedAccountKeeper, msg types.MsgSetReferee) sdk.Error {
	senderAcc := ak.GetAccount(ctx, msg.Sender)
	if senderAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("sender %s is not exist yet", msg.Sender))
	}
	RefereeAcc := ak.GetAccount(ctx, msg.Referee)
	if RefereeAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("referee %s is not exist yet", msg.Referee))
	}

	RefereeAccx := k.GetOrCreateAccountX(ctx, msg.Referee)
	if RefereeAccx.MemoRequired {
		return types.ErrRefereeMemoRequired(msg.Referee.String())
	}

	if k.BlacklistedAddr(msg.Referee) {
		return sdk.ErrInvalidAddress("referee can not be module address")
	}
	return nil
}
func preCheckTime(timeNow int64, refereeChangeTime int64, refereeChangeMinInterval int64, referee sdk.AccAddress) sdk.Error {

	if timeNow-refereeChangeTime < refereeChangeMinInterval {
		return types.ErrRefereeChangeTooFast(referee.String())
	}
	return nil
}
