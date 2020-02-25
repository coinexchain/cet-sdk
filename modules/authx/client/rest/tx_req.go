package rest

import (
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/coinexchain/cosmos-utils/client/restutil"

	"github.com/coinexchain/cet-sdk/modules/authx/internal/types"
)

type setRefereeReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Referee string       `json:"referee"`
}

func (req *setRefereeReq) New() restutil.RestReq {
	return new(setRefereeReq)
}
func (req *setRefereeReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}
func (req *setRefereeReq) GetMsg(r *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	referee, err := sdk.AccAddressFromBech32(req.Referee)
	if err != nil {
		return nil, err
	}
	return types.NewMsgSetReferee(sender, referee), nil
}
