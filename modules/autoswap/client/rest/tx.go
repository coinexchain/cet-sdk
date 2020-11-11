package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"net/http"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cosmos-utils/client/restutil"
)

/* addLiquidityReq */

type addLiquidityReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Stock       string       `json:"stock"`
	Money       string       `json:"money"`
	NoSwap      bool         `json:"no_swap"`
	NoOrderBook bool         `json:"no_order_book"`
	StockIn     string       `json:"stock_in"`
	MoneyIn     string       `json:"money_in"`
	To          string       `json:"to"`
}

func (req *addLiquidityReq) New() restutil.RestReq {
	return new(addLiquidityReq)
}

func (req *addLiquidityReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}

func (req *addLiquidityReq) GetMsg(_ *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	msg := &types.MsgAddLiquidity{
		Sender: sender,
		Stock:  req.Stock,
		Money:  req.Money,
	}

	var err error
	if msg.StockIn, err = parseSdkInt("init_stock", req.StockIn); err != nil {
		return nil, err
	}
	if msg.MoneyIn, err = parseSdkInt("init_money", req.MoneyIn); err != nil {
		return nil, err
	}
	if msg.To, err = sdk.AccAddressFromBech32(req.To); err != nil {
		return nil, err
	}

	return msg, err
}

/* removeLiquidityReq */

type removeLiquidityReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Stock       string       `json:"stock"`
	Money       string       `json:"money"`
	NoSwap      bool         `json:"no_swap"`
	NoOrderBook bool         `json:"no_order_book"`
	StockMin    string       `json:"stock_min"`
	MoneyMin    string       `json:"money_min"`
	Amount      string       `json:"amount"`
	To          string       `json:"to"`
}

func (req *removeLiquidityReq) New() restutil.RestReq {
	return new(removeLiquidityReq)
}

func (req *removeLiquidityReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}

func (req *removeLiquidityReq) GetMsg(_ *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	msg := &types.MsgRemoveLiquidity{
		Sender: sender,
		Stock:  req.Stock,
		Money:  req.Money,
	}

	var err error
	if msg.Amount, err = parseSdkInt("amount", req.Amount); err != nil {
		return nil, err
	}
	if msg.To, err = sdk.AccAddressFromBech32(req.To); err != nil {
		return nil, err
	}

	return msg, err
}

/* createHandlerFns */
func addLiquidityHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	var req addLiquidityReq
	return restutil.NewRestHandler(cdc, cliCtx, &req)
}
func removeLiquidityHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	var req removeLiquidityReq
	return restutil.NewRestHandler(cdc, cliCtx, &req)
}

/* helpers */

func parseSdkInt(name, s string) (val sdk.Int, err error) {
	ok := false
	if val, ok = sdk.NewIntFromString(s); !ok {
		err = fmt.Errorf("%s must be a valid integer number", name)
	}
	return
}
