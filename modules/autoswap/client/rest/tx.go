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
		Owner:           sender,
		Stock:           req.Stock,
		Money:           req.Money,
		IsSwapOpen:      !req.NoSwap,
		IsOrderBookOpen: !req.NoOrderBook,
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

/* addLiquidityReq */

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
		Sender:          sender,
		Stock:           req.Stock,
		Money:           req.Money,
		IsSwapOpen:      !req.NoSwap,
		IsOrderBookOpen: !req.NoOrderBook,
	}

	var err error
	if msg.Amount, err = parseSdkInt("amount", req.Amount); err != nil {
		return nil, err
	}
	if msg.AmountStockMin, err = parseSdkInt("stock_min", req.StockMin); err != nil {
		return nil, err
	}
	if msg.AmountMoneyMin, err = parseSdkInt("money_min", req.MoneyMin); err != nil {
		return nil, err
	}
	if msg.To, err = sdk.AccAddressFromBech32(req.To); err != nil {
		return nil, err
	}

	return msg, err
}

/* swapTokensReq */

type pairInfo struct {
	Symbol      string `json:"pair"`
	NoSwap      bool   `json:"no_swap"`
	NoOrderBook bool   `json:"no_order_book"`
}
type swapTokensReq struct {
	BaseReq   rest.BaseReq `json:"base_req"`
	Path      []pairInfo   `json:"path"`
	Side      string       `json:"side"`
	Amount    string       `json:"amount"`
	OutputMin string       `json:"output_min"`
}

func (req *swapTokensReq) New() restutil.RestReq {
	return new(swapTokensReq)
}

func (req *swapTokensReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}

func (req *swapTokensReq) GetMsg(_ *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	msg := &types.MsgSwapTokens{
		Sender: sender,
	}
	for _, pairInfo := range req.Path {
		msg.Pairs = append(msg.Pairs, types.MarketInfo{
			MarketSymbol:    pairInfo.Symbol,
			IsOpenSwap:      !pairInfo.NoSwap,
			IsOpenOrderBook: !pairInfo.NoOrderBook,
		})
	}

	var err error
	if msg.IsBuy, err = parseIsBuy(req.Side); err != nil {
		return nil, err
	}
	if msg.Amount, err = parseSdkInt("amount", req.Amount); err != nil {
		return nil, err
	}
	if msg.MinOutputAmount, err = parseSdkInt("output_min", req.OutputMin); err != nil {
		return nil, err
	}

	return msg, err
}

/* createLimitOrderReq */

type createLimitOrderReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	PairSymbol  string       `json:"pair"`
	NoSwap      bool         `json:"no-swap"`
	NoOrderBook bool         `json:"no_order_book"`
	Side        string       `json:"side"`
	Amount      string       `json:"amount"`
	OrderID     string       `json:"order_id"`
	Price       string       `json:"price"`
	PrevKey     string       `json:"prev_key"`
}

func (req *createLimitOrderReq) New() restutil.RestReq {
	return new(createLimitOrderReq)
}

func (req *createLimitOrderReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}

func (req *createLimitOrderReq) GetMsg(_ *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	msg := &types.MsgCreateLimitOrder{
		OrderBasic: types.OrderBasic{
			Sender:          sender,
			MarketSymbol:    req.PairSymbol,
			IsOpenSwap:      !req.NoSwap,
			IsOpenOrderBook: !req.NoOrderBook,
			IsLimitOrder:    true,
		},
	}
	var err error
	if msg.IsBuy, err = parseIsBuy(req.Side); err != nil {
		return nil, err
	}
	if msg.Amount, err = parseSdkInt("amount", req.Amount); err != nil {
		return nil, err
	}
	if msg.Price, err = parseSdkDec("price", req.Price); err != nil {
		return nil, err
	}
	if msg.OrderID, err = parseInt64("order_id", req.OrderID); err != nil {
		return nil, err
	}
	// TODO: prevKey

	return msg, err
}

/* cancelOrderReq */

type cancelOrderReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	PairSymbol  string       `json:"pair"`
	NoSwap      bool         `json:"no-swap"`
	NoOrderBook bool         `json:"no_order_book"`
	Side        string       `json:"side"`
	OrderID     string       `json:"order_id"`
	PrevKey     string       `json:"prev_key"`
}

func (req *cancelOrderReq) New() restutil.RestReq {
	return new(cancelOrderReq)
}

func (req *cancelOrderReq) GetBaseReq() *rest.BaseReq {
	return &req.BaseReq
}

func (req *cancelOrderReq) GetMsg(_ *http.Request, sender sdk.AccAddress) (sdk.Msg, error) {
	msg := &types.MsgDeleteOrder{
		Sender:          sender,
		MarketSymbol:    req.PairSymbol,
		IsOpenSwap:      !req.NoSwap,
		IsOpenOrderBook: !req.NoOrderBook,
	}
	var err error
	if msg.IsBuy, err = parseIsBuy(req.Side); err != nil {
		return nil, err
	}
	if msg.OrderID, err = parseInt64("order_id", req.OrderID); err != nil {
		return nil, err
	}
	// TODO: prevKey

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
func swapTokensHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	var req swapTokensReq
	return restutil.NewRestHandler(cdc, cliCtx, &req)
}
func createLimitOrderHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	var req createLimitOrderReq
	return restutil.NewRestHandler(cdc, cliCtx, &req)
}
func cancelOrderHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	var req cancelOrderReq
	return restutil.NewRestHandler(cdc, cliCtx, &req)
}

/* helpers */

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

func parseInt64(name, s string) (val int64, err error) {
	if val, err = strconv.ParseInt(s, 10, 64); err != nil {
		err = fmt.Errorf("%s must be a valid integer number", name)
	}
	return
}

func parseSdkInt(name, s string) (val sdk.Int, err error) {
	ok := false
	if val, ok = sdk.NewIntFromString(s); !ok {
		err = fmt.Errorf("%s must be a valid integer number", name)
	}
	return
}
func parseSdkDec(name, s string) (val sdk.Dec, err error) {
	if val, err = sdk.NewDecFromStr(s); err != nil {
		err = fmt.Errorf("%s must be a valid decimal number", name)
	}
	return
}
