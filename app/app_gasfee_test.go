package app

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/dex/modules/asset"
	"github.com/coinexchain/dex/modules/authx"
	"github.com/coinexchain/dex/modules/authx/types"
	"github.com/coinexchain/dex/modules/bankx"
	"github.com/coinexchain/dex/modules/market"
	"github.com/coinexchain/dex/testutil"
	dex "github.com/coinexchain/dex/types"
)

func TestGasFeeDeductedWhenTxFailed(t *testing.T) {
	// acc & app
	key, acc := testutil.NewBaseAccount(10000000000, 0, 0)
	app := initAppWithBaseAccounts(acc)

	// begin block
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// deliver tx
	coins := dex.NewCetCoins(100000000000)
	toAddr := sdk.AccAddress([]byte("addr"))
	msg := bankx.NewMsgSend(acc.Address, toAddr, coins, 0)
	tx := newStdTxBuilder().
		Msgs(msg).GasAndFee(1000000, 100).AccNumSeqKey(0, 0, key).Build()

	result := app.Deliver(tx)
	require.Equal(t, sdk.CodeInsufficientCoins, result.Code)

	// end block & commit
	app.EndBlock(abci.RequestEndBlock{Height: 1})
	app.Commit()

	// check coins
	ctx := app.NewContext(true, abci.Header{})
	require.Equal(t, int64(10000000000-100),
		app.accountKeeper.GetAccount(ctx, acc.Address).GetCoins().AmountOf("cet").Int64())
}

func TestMinGasPriceLimit(t *testing.T) {
	// acc & app
	key, acc := testutil.NewBaseAccount(1e10, 0, 0)
	app := initAppWithBaseAccounts(acc)

	// begin block
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// deliver tx
	coins := dex.NewCetCoins(1e8)
	toAddr := sdk.AccAddress([]byte("addr"))
	msg := bankx.NewMsgSend(acc.Address, toAddr, coins, 0)
	tx := newStdTxBuilder().
		Msgs(msg).GasAndFee(10000000000, 1).AccNumSeqKey(0, 0, key).Build()

	result := app.Deliver(tx)
	require.Equal(t, types.CodeGasPriceTooLow, result.Code)
}

func TestSmallAccountGasCost(t *testing.T) {
	// acc & app
	key, acc := testutil.NewBaseAccount(1e10, 0, 0)
	app := initAppWithBaseAccounts(acc)

	// begin block
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// deliver tx
	coins := dex.NewCetCoins(1e8)
	toAddr := sdk.AccAddress([]byte("addr"))
	msg := bankx.NewMsgSend(acc.Address, toAddr, coins, 0)
	tx := newStdTxBuilder().
		Msgs(msg).GasAndFee(90000, 100).AccNumSeqKey(0, 0, key).Build()

	// ok
	result := app.Deliver(tx)
	require.Equal(t, sdk.CodeOK, result.Code)
	require.Equal(t, 90000, int(result.GasWanted))
	require.Equal(t, 57371, int(result.GasUsed))
}

func TestBigAccountGasCost(t *testing.T) {
	// acc & app
	key, acc := testutil.NewBaseAccount(1e10, 0, 0)
	for i := 0; i < 1000; i++ {
		coin := sdk.NewCoin(fmt.Sprintf("coin%d", i), sdk.NewInt(1e10))
		acc.Coins = acc.Coins.Add(sdk.NewCoins(coin))
	}
	app := initAppWithBaseAccounts(acc)

	// begin block
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// deliver tx
	coins := dex.NewCetCoins(1e8)
	toAddr := sdk.AccAddress([]byte("addr"))
	msg := bankx.NewMsgSend(acc.Address, toAddr, coins, 0)
	tx := newStdTxBuilder().
		Msgs(msg).GasAndFee(9000000, 100).AccNumSeqKey(0, 0, key).Build()

	// ok
	result := app.Deliver(tx)
	require.Equal(t, sdk.CodeOK, result.Code)
	require.Equal(t, 9000000, int(result.GasWanted))
	require.Equal(t, 3569201, int(result.GasUsed))
}

func TestBigAuthxAccountCreateOrderGasCost(t *testing.T) {
	// acc & app
	key, acc := testutil.NewBaseAccount(1e16, 0, 0)
	_, acc2 := testutil.NewBaseAccount(1e8, 1, 0)

	for i := 0; i < 1000; i++ {
		coin := sdk.NewCoin(fmt.Sprintf("coin%d", i), sdk.NewInt(1e10))
		acc2.Coins = acc2.Coins.Add(sdk.NewCoins(coin))
	}
	app := initAppWithAccounts(acc, acc2)

	// begin block
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := app.NewContext(false, header)

	var (
		stock             = "usdt000"
		money             = "cet"
		issueAmount int64 = 210000000000
	)

	// issue tokens
	msgStock := asset.NewMsgIssueToken(stock, stock, sdk.NewInt(issueAmount), acc.Address,
		false, false, false, false, "", "", "")
	tx := newStdTxBuilder().
		Msgs(msgStock).GasAndFee(9000000, 100).AccNumSeqKey(0, 0, key).Build()
	res := app.Deliver(tx)
	require.Equal(t, sdk.CodeOK, res.Code)

	//create market info
	msgMarketInfo := market.MsgCreateTradingPair{Stock: stock, Money: money, Creator: acc.Address, PricePrecision: 8}
	tx = newStdTxBuilder().
		Msgs(msgMarketInfo).GasAndFee(9000000, 100).AccNumSeqKey(0, 1, key).Build()
	res = app.Deliver(tx)
	require.Equal(t, sdk.CodeOK, res.Code)

	for i := 0; i < 1000; i++ {
		coin := sdk.NewCoin(fmt.Sprintf("coin%d", i), sdk.NewInt(1e8))
		_ = app.supplyKeeper.SendCoinsFromAccountToModule(ctx, acc2.Address, authx.ModuleName, sdk.Coins{coin})
	}

	//create trading pair
	msgCreateOrder := market.MsgCreateOrder{
		Sender:         acc.Address,
		Sequence:       2,
		TradingPair:    stock + market.SymbolSeparator + money,
		OrderType:      market.LimitOrder,
		PricePrecision: 8,
		Price:          100,
		Quantity:       10000000,
		Side:           market.SELL,
		TimeInForce:    market.GTE,
	}
	tx = newStdTxBuilder().
		Msgs(msgCreateOrder).GasAndFee(9000000, 100).AccNumSeqKey(0, 2, key).Build()

	// ok
	result := app.Deliver(tx)
	require.Equal(t, sdk.CodeOK, result.Code)
	require.Equal(t, 9000000, int(result.GasWanted))
	require.Equal(t, 83386, int(result.GasUsed))
}

func TestSmallAuthxAccountCreateOrderGasCost(t *testing.T) {
	// acc & app
	key, acc := testutil.NewBaseAccount(1e16, 0, 0)

	app := initAppWithAccounts(acc)

	// begin block
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	var (
		stock             = "usdt000"
		money             = "cet"
		issueAmount int64 = 210000000000
	)

	// issue tokens
	msgStock := asset.NewMsgIssueToken(stock, stock, sdk.NewInt(issueAmount), acc.Address,
		false, false, false, false, "", "", "")
	tx := newStdTxBuilder().
		Msgs(msgStock).GasAndFee(9000000, 100).AccNumSeqKey(0, 0, key).Build()
	res := app.Deliver(tx)
	require.Equal(t, sdk.CodeOK, res.Code)

	//create market info
	msgMarketInfo := market.MsgCreateTradingPair{Stock: stock, Money: money, Creator: acc.Address, PricePrecision: 8}
	tx = newStdTxBuilder().
		Msgs(msgMarketInfo).GasAndFee(9000000, 100).AccNumSeqKey(0, 1, key).Build()
	res = app.Deliver(tx)
	require.Equal(t, sdk.CodeOK, res.Code)

	//create trading pair
	msgCreateOrder := market.MsgCreateOrder{
		Sender:         acc.Address,
		Sequence:       2,
		TradingPair:    stock + market.SymbolSeparator + money,
		OrderType:      market.LimitOrder,
		PricePrecision: 8,
		Price:          100,
		Quantity:       10000000,
		Side:           market.SELL,
		TimeInForce:    market.GTE,
	}
	tx = newStdTxBuilder().
		Msgs(msgCreateOrder).GasAndFee(9000000, 100).AccNumSeqKey(0, 2, key).Build()

	// ok
	result := app.Deliver(tx)
	require.Equal(t, sdk.CodeOK, result.Code)
	require.Equal(t, 9000000, int(result.GasWanted))
	require.Equal(t, 83386, int(result.GasUsed))
}
