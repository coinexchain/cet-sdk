package authx

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/authx/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/authx/internal/types"
	dex "github.com/coinexchain/cet-sdk/types"
)

func NewHandler(k keepers.AccountXKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgSetReferee:
			return handleMsgSetReferee(ctx, k, msg)
		default:
			return dex.ErrUnknownRequest(ModuleName, msg)

		}
	}
}
func handleMsgSetReferee(ctx sdk.Context, k keepers.AccountXKeeper, msg types.MsgSetReferee) sdk.Result {

	accx, exist := k.GetAccountX(ctx, msg.Sender)
	if !exist {
		return sdk.ErrUnknownAddress(fmt.Sprintf("%s is not exist yet", msg.Sender)).Result()
	}
	refereeMinChangeInterval := k.GetParams(ctx).RefereeChangeMinInterval
	err := accx.UpdateRefereeAddr(msg.Referee, ctx.BlockTime().UnixNano(), refereeMinChangeInterval)
	if err != nil {
		return err.Result()
	}
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
