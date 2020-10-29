package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/coinexchain/cosmos-utils/client/restutil"
)

/* addLiquidityReq */

type addLiquidityReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Stock   string       `json:"stock"`
	Money   string       `json:"money"`
	StockIn string       `json:"stock_in"`
	MoneyIn string       `json:"money_in"`
	NoSwap  bool         `json:"no_swap"`
	To      string       `json:"to"`
}

func (req *addLiquidityReq) New() restutil.RestReq {
	return new(addLiquidityReq)
}

func (req *addLiquidityReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}

func (req *addLiquidityReq) GetMsg(r *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	msg := &types.MsgAddLiquidity{
		Owner:      sender,
		Stock:      req.Stock,
		Money:      req.Money,
		IsOpenSwap: !req.NoSwap,
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

/* createMarketOrderReq */

type createMarketOrderReq struct {
	BaseReq    rest.BaseReq `json:"base_req"`
	PairSymbol string       `json:"pair"`
	NoSwap     bool         `json:"no-swap"`
	Side       string       `json:"side"`
	Amount     string       `json:"amount"`
}

func (req *createMarketOrderReq) New() restutil.RestReq {
	return new(createMarketOrderReq)
}

func (req *createMarketOrderReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}

func (req *createMarketOrderReq) GetMsg(r *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	msg := &types.MsgCreateMarketOrder{
		OrderBasic: types.OrderBasic{
			Sender:       sender,
			MarketSymbol: req.PairSymbol,
			IsOpenSwap:   !req.NoSwap,
			IsLimitOrder: false,
		},
	}
	var err error
	if msg.IsBuy, err = parseIsBuy(req.Side); err != nil {
		return nil, err
	}
	if msg.Amount, err = strconv.ParseInt(req.Amount, 10, 64); err != nil {
		return nil, errors.New("invalid amount")
	}

	return msg, err
}

/* createLimitOrderReq */

type createLimitOrderReq struct {
	BaseReq        rest.BaseReq `json:"base_req"`
	PairSymbol     string       `json:"pair"`
	NoSwap         bool         `json:"no-swap"`
	Side           string       `json:"side"`
	Amount         string       `json:"amount"`
	OrderID        string       `json:"order_id"`
	Price          string       `json:"price"`
	PricePrecision byte         `json:"price_precision"`
	PrevKey        string       `json:"prev_key"`
}

func (req *createLimitOrderReq) New() restutil.RestReq {
	return new(createLimitOrderReq)
}

func (req *createLimitOrderReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}

func (req *createLimitOrderReq) GetMsg(r *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	msg := &types.MsgCreateLimitOrder{
		OrderBasic: types.OrderBasic{
			Sender:       sender,
			MarketSymbol: req.PairSymbol,
			IsOpenSwap:   !req.NoSwap,
			IsLimitOrder: true,
		},
		PricePrecision: req.PricePrecision,
	}
	var err error
	if msg.IsBuy, err = parseIsBuy(req.Side); err != nil {
		return nil, err
	}
	if msg.Amount, err = strconv.ParseInt(req.Amount, 10, 64); err != nil {
		return nil, errors.New("invalid amount")
	}
	if msg.OrderID, err = strconv.ParseInt(req.OrderID, 10, 64); err != nil {
		return nil, errors.New("invalid order_id")
	}
	if msg.Price, err = strconv.ParseInt(req.Price, 10, 64); err != nil {
		return nil, errors.New("invalid price")
	}
	// TODO: prevKey

	return msg, err
}

/* createHandlerFns */
func addLiquidityHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	var req addLiquidityReq
	return restutil.NewRestHandler(cdc, cliCtx, &req)
}
func createMarketOrderHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	var req createMarketOrderReq
	return restutil.NewRestHandler(cdc, cliCtx, &req)
}
func createLimitOrderHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	var req createLimitOrderReq
	return restutil.NewRestHandler(cdc, cliCtx, &req)
}

/* helpers */

func parseSdkInt(name, s string) (val sdk.Int, err error) {
	ok := false
	if val, ok = sdk.NewIntFromString(s); !ok {
		err = fmt.Errorf("%s must be a valid integer", name)
	}
	return
}

func parseIsBuy(side string) (bool, error) {
	side = strings.ToLower(side)
	if side == "buy" {
		return true, nil
	}
	if side == "sell" {
		return false, nil
	}
	return false, errors.New("side must be buy or sell")
}
