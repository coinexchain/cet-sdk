package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/coinexchain/cet-sdk/modules/authx"
	simulation2 "github.com/coinexchain/cet-sdk/simulation"
)

func SimulateMsgSetReferee(k authx.AccountXKeeper, ak authx.ExpectedAccountKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simulation.Account) (
		OperationMsg simulation.OperationMsg, futureOps []simulation.FutureOperation, err error) {

		sender := simulation.RandomAcc(r, accounts)
		referee := simulation.RandomAcc(r, accounts)
		for sender.Address.Equals(referee.Address) {
			referee = simulation.RandomAcc(r, accounts)
		}

		msg := authx.MsgSetReferee{
			Sender:  sender.Address,
			Referee: referee.Address,
		}
		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(authx.ModuleName), nil, nil
		}

		ok := simulation2.SimulateHandleMsg(msg, authx.NewHandler(k, ak), ctx)

		opMsg := simulation.NewOperationMsg(msg, ok, "")
		if !ok {
			return opMsg, nil, nil
		}
		ok = checkSetReferee(ctx, k, msg)
		if !ok {
			return simulation.NewOperationMsg(msg, ok, ""), nil, fmt.Errorf("set referee failed")
		}
		return opMsg, nil, nil
	}
}
func checkSetReferee(ctx sdk.Context, k authx.AccountXKeeper, msg authx.MsgSetReferee) bool {
	senderAcc, ok := k.GetAccountX(ctx, msg.Sender)
	return ok && senderAcc.Referee.Equals(msg.Referee) && senderAcc.RefereeChangeTime > 0
}
