package keepers

import (
	"github.com/coinexchain/cet-sdk/modules/autoswap/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
)

var _ IPairKeeper = &PairKeeper{}

type Pair struct {
	Symbol string
}

type IPairKeeper interface {
	IPoolKeeper
	AddLimitOrder(ctx sdk.Context, order *types.Order) sdk.Error
	DeleteOrder(ctx sdk.Context, order *types.MsgCancelOrder) sdk.Error
	HasOrder(ctx sdk.Context, orderID string) bool
	GetOrder(ctx sdk.Context, orderID string) *types.Order

	SetParams(ctx sdk.Context, params types.Params)
	GetParams(ctx sdk.Context) types.Params
	GetTakerFee(ctx sdk.Context) sdk.Dec
	GetMakerFee(ctx sdk.Context) sdk.Dec
	GetDealWithPoolFee(ctx sdk.Context) sdk.Dec
	GetFeeToValidator(ctx sdk.Context) sdk.Dec

	GetPairList() map[Pair]struct{}
	ClearPairList()
}

type FeeFunc func(sdk.Context) sdk.Dec

type PairKeeper struct {
	IPoolKeeper
	IOrderBookKeeper
	types.ExpectedAccountKeeper
	types.SupplyKeeper
	types.ExpectedBankKeeper
	codec    *codec.Codec
	storeKey sdk.StoreKey
	subspace params.Subspace
	// record deal pairs in one block.
	DealPairs map[Pair]struct{}
}

func NewPairKeeper(poolKeeper IPoolKeeper, supplyK types.SupplyKeeper, bnk types.ExpectedBankKeeper,
	codec *codec.Codec, storeKey sdk.StoreKey, paramSubspace params.Subspace) *PairKeeper {
	return &PairKeeper{
		codec:              codec,
		storeKey:           storeKey,
		IPoolKeeper:        poolKeeper,
		SupplyKeeper:       supplyK,
		ExpectedBankKeeper: bnk,
		subspace:           paramSubspace.WithKeyTable(types.ParamKeyTable()),
		IOrderBookKeeper: &OrderKeeper{
			codec:    codec,
			storeKey: storeKey,
		},
		DealPairs: make(map[Pair]struct{}),
	}
}

func (pk *PairKeeper) GetPairList() map[Pair]struct{} {
	return pk.DealPairs
}
func (pk *PairKeeper) ClearPairList() {
	pk.DealPairs = make(map[Pair]struct{})
}

func (pk *PairKeeper) SetParams(ctx sdk.Context, params types.Params) {
	pk.subspace.SetParamSet(ctx, &params)
}

func (pk *PairKeeper) GetParams(ctx sdk.Context) (param types.Params) {
	pk.subspace.GetParamSet(ctx, &param)
	return
}

func (pk *PairKeeper) GetTakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(pk.GetParams(ctx).TakerFeeRateRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetMakerFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(pk.GetParams(ctx).MakerFeeRateRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetDealWithPoolFee(ctx sdk.Context) sdk.Dec {
	return sdk.NewDec(pk.GetParams(ctx).DealWithPoolFeeRate).QuoInt64(types.DefaultFeePrecision)
}

func (pk *PairKeeper) GetFeeToValidator(ctx sdk.Context) sdk.Dec {
	param := pk.GetParams(ctx)
	return sdk.NewDec(param.FeeToValidator).QuoInt64(param.FeeToValidator + param.FeeToPool)
}

func (pk *PairKeeper) AllocateFeeToValidatorAndPool(ctx sdk.Context, denom string, totalAmount sdk.Int, sender sdk.AccAddress) sdk.Error {
	feeToVal := pk.GetFeeToValidator(ctx).MulInt(totalAmount).TruncateInt()
	feeToPool := totalAmount.Sub(feeToVal)
	err := pk.SendCoinsFromAccountToModule(ctx, sender, auth.FeeCollectorName, sdk.NewCoins(sdk.NewCoin(denom, feeToVal)))
	if err != nil {
		return err
	}
	err = pk.SendCoinsFromAccountToModule(ctx, sender, types.PoolModuleAcc, sdk.NewCoins(sdk.NewCoin(denom, feeToPool)))
	if err != nil {
		return err
	}
	return nil
}

func (pk *PairKeeper) AllocateFeeToValidator(ctx sdk.Context, fee sdk.Coins, sender sdk.AccAddress) sdk.Error {
	err := pk.SendCoinsFromAccountToModule(ctx, sender, auth.FeeCollectorName, fee)
	if err != nil {
		return err
	}
	return nil
}

func (pk PairKeeper) AddLimitOrder(ctx sdk.Context, order *types.Order) (err sdk.Error) {
	defer func() {
		r := recover()
		switch r.(type) {
		case sdk.Error:
			err = r.(sdk.Error)
		case string:
			err = sdk.NewError(types.RouterKey, types.CodeUnKnownError, r.(string))
		case error:
			err = sdk.NewError(types.RouterKey, types.CodeUnKnownError, r.(error).Error())
		}
	}()

	order.Sequence = int64(pk.GetAccount(ctx, order.Sender).GetSequence())
	actualAmount, err := pk.freezeOrderCoin(ctx, order)
	if err != nil {
		return err
	}
	dealInfo := &types.DealInfo{
		RemainAmount:    actualAmount,
		AmountInToPool:  sdk.ZeroInt(),
		DealMoneyInBook: sdk.ZeroInt(),
		DealStockInBook: sdk.ZeroInt(),
	}

	// 1. get matched opposite orders.
	oppositeOrders := pk.GetMatchedOrder(ctx, order)

	// 2. deal order with pool
	poolInfo := pk.GetPoolInfo(ctx, order.TradingPair)
	for _, opOrder := range oppositeOrders {
		if allDeal, err := pk.dealOrderWithOrderBookAndPool(ctx, order, opOrder, dealInfo, poolInfo); allDeal {
			break
		} else if err != nil {
			return err
		}
	}

	// 3. final deal with pool and order
	pk.finalDealWithPool(ctx, order, dealInfo, poolInfo)

	// 4. store order in keeper if need
	pk.storeOrderIfNeed(ctx, order)
	return nil
}

func (pk PairKeeper) freezeOrderCoin(ctx sdk.Context, order *types.Order) (sdk.Int, sdk.Error) {
	orderAmount := order.ActualAmount()
	if order.IsBuy {
		if err := pk.FreezeCoins(ctx, order.Sender, newCoins(order.Money(), orderAmount)); err != nil {
			return sdk.Int{}, err
		}
	} else {
		if err := pk.FreezeCoins(ctx, order.Sender, newCoins(order.Stock(), orderAmount)); err != nil {
			return sdk.Int{}, err
		}
	}
	return orderAmount, nil
}

func (pk PairKeeper) dealOrderWithOrderBookAndPool(ctx sdk.Context, order,
	oppositeOrder *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) (allDeal bool, err sdk.Error) {
	if poolInfo != nil {
		pk.tryDealInPool(dealInfo, oppositeOrder.Price, order.IsBuy, poolInfo)
	}
	pk.dealInOrderBook(ctx, order, oppositeOrder, dealInfo, poolInfo != nil)
	if oppositeOrder.LeftStock == 0 {
		pk.DelOrder(ctx, order)
	}
	return order.LeftStock == 0, nil
}

func (pk PairKeeper) tryDealInPool(dealInfo *types.DealInfo, dealPrice sdk.Dec, isBuy bool, info *PoolInfo) bool {
	currTokenCanTradeWithPool := pk.intoPoolAmountTillPrice(dealPrice, isBuy, info)
	// will check deal token amount later.
	if currTokenCanTradeWithPool.GT(types.MaxAmount) {
		panic("deal amount with pool is too large")
	}
	if currTokenCanTradeWithPool.GT(dealInfo.AmountInToPool) {
		diffTokenTradeWithPool := currTokenCanTradeWithPool.Sub(dealInfo.AmountInToPool)
		allDeal := diffTokenTradeWithPool.GT(dealInfo.RemainAmount)
		if allDeal {
			diffTokenTradeWithPool = dealInfo.RemainAmount
		}
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(diffTokenTradeWithPool)
		dealInfo.AmountInToPool = dealInfo.AmountInToPool.Add(diffTokenTradeWithPool)
		return allDeal
	}
	return false
}

func (pk PairKeeper) intoPoolAmountTillPrice(dealPrice sdk.Dec, isBuy bool, info *PoolInfo) sdk.Int {
	if isBuy {
		root := dealPrice.Mul(sdk.NewDecFromInt(info.StockAmmReserve)).Mul(sdk.NewDecFromInt(info.MoneyAmmReserve))
		return sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.Int)).Sub(sdk.NewDecFromInt(info.MoneyAmmReserve)).TruncateInt()
	}
	root := sdk.NewDecFromInt(info.StockAmmReserve).Mul(sdk.NewDecFromInt(info.MoneyAmmReserve)).Quo(dealPrice)
	return sdk.NewDecFromBigInt(sdk.NewDec(0).Sqrt(root.Int)).Sub(sdk.NewDecFromInt(info.StockAmmReserve)).TruncateInt()
}

func (pk PairKeeper) dealInOrderBook(ctx sdk.Context, currOrder, orderInBook *types.Order, dealInfo *types.DealInfo, isPoolExists bool) {
	// calculate stock amount
	dealStockAmount := currOrder.LeftStock
	if orderInBook.LeftStock < currOrder.LeftStock {
		dealStockAmount = orderInBook.LeftStock
	}
	currOrder.LeftStock -= dealStockAmount
	orderInBook.LeftStock -= dealStockAmount

	var (
		stockFee   sdk.Int
		moneyFee   sdk.Int
		stockTrans = sdk.NewInt(dealStockAmount)
		moneyTrans = sdk.NewDec(dealStockAmount).Mul(orderInBook.Price).TruncateInt()
	)
	if currOrder.IsBuy {
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(moneyTrans)
		moneyFee = pk.GetMakerFee(ctx).MulInt(moneyTrans).TruncateInt()
		stockFee = pk.GetTakerFee(ctx).MulInt(stockTrans).TruncateInt()
	} else {
		dealInfo.RemainAmount = dealInfo.RemainAmount.Sub(stockTrans)
		stockFee = pk.GetMakerFee(ctx).MulInt(stockTrans).TruncateInt()
		moneyFee = pk.GetTakerFee(ctx).MulInt(moneyTrans).TruncateInt()
	}

	dealInfo.DealMoneyInBook = dealInfo.DealMoneyInBook.Add(moneyTrans)
	dealInfo.DealStockInBook = dealInfo.DealStockInBook.Add(stockTrans)
	// transfer tokens in orders
	if currOrder.IsBuy {
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Money(), moneyTrans.Sub(moneyFee))
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Stock(), stockTrans.Sub(stockFee))
		if isPoolExists {
			if err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Money(), moneyFee, currOrder.Sender); err != nil {
				panic(err)
			}
			if err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Stock(), stockFee, orderInBook.Sender); err != nil {
				panic(err)
			}
		} else {
			if err := pk.AllocateFeeToValidator(ctx, sdk.NewCoins(sdk.NewCoin(currOrder.Money(), moneyFee),
				sdk.NewCoin(currOrder.Stock(), stockFee)), currOrder.Sender); err != nil {
				panic(err)
			}
		}
	} else {
		pk.transferToken(ctx, currOrder.Sender, orderInBook.Sender, currOrder.Stock(), stockTrans.Sub(stockFee))
		pk.transferToken(ctx, orderInBook.Sender, currOrder.Sender, currOrder.Money(), moneyTrans.Sub(moneyFee))
		if err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Stock(), stockFee, currOrder.Sender); err != nil {
			panic(err)
		}
		if err := pk.AllocateFeeToValidatorAndPool(ctx, currOrder.Money(), moneyFee, orderInBook.Sender); err != nil {
			panic(err)
		}
	}
}

func (pk PairKeeper) finalDealWithPool(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) {
	pk.dealWithPoolAndCollectFee(ctx, order, dealInfo, poolInfo)
	if dealInfo.AmountInToPool.IsPositive() {
		// todo emit deal with pool log
	}
	pk.SetPoolInfo(ctx, order.TradingPair, poolInfo)
}

func (pk PairKeeper) dealWithPoolAndCollectFee(ctx sdk.Context, order *types.Order, dealInfo *types.DealInfo, poolInfo *PoolInfo) sdk.Int {
	outPoolTokenReserve, inPoolTokenReserve, otherToTaker := poolInfo.MoneyAmmReserve, poolInfo.StockAmmReserve, dealInfo.DealMoneyInBook
	if order.IsBuy {
		outPoolTokenReserve, inPoolTokenReserve, otherToTaker = poolInfo.StockAmmReserve, poolInfo.MoneyAmmReserve, dealInfo.DealStockInBook
	}
	outAmount := sdk.ZeroInt()
	if !dealInfo.AmountInToPool.IsZero() {
		outAmount = outPoolTokenReserve.Mul(dealInfo.AmountInToPool).Quo(inPoolTokenReserve.Add(dealInfo.AmountInToPool))
	} else {
		return sdk.ZeroInt()
	}

	// add fee calculate
	fee := pk.GetDealWithPoolFee(ctx).MulInt(outAmount).TruncateInt()
	amountToTaker := outAmount.Add(otherToTaker).Sub(fee)
	outAmount = outAmount.Sub(fee)
	if order.IsBuy {
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Sub(outAmount)
	} else {
		poolInfo.StockAmmReserve = poolInfo.StockAmmReserve.Add(dealInfo.AmountInToPool)
		poolInfo.MoneyAmmReserve = poolInfo.MoneyAmmReserve.Sub(outAmount)
	}
	// transfer token from pool to order sender
	if order.IsBuy {
		if err := pk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, order.Sender, newCoins(order.Stock(), outAmount)); err != nil {
			panic(err)
		}
		if err := pk.AllocateFeeToValidatorAndPool(ctx, order.Money(), dealInfo.AmountInToPool, order.Sender); err != nil {
			panic(err)
		}
	} else {
		if err := pk.SendCoinsFromModuleToAccount(ctx, types.PoolModuleAcc, order.Sender, newCoins(order.Money(), outAmount)); err != nil {
			panic(err)
		}
		if err := pk.AllocateFeeToValidatorAndPool(ctx, order.Stock(), dealInfo.AmountInToPool, order.Sender); err != nil {
			panic(err)
		}
	}
	return amountToTaker
}

func (pk PairKeeper) storeOrderIfNeed(ctx sdk.Context, order *types.Order) {
	if order.LeftStock > 0 {
		pk.AddOrder(ctx, order)
	}
}

func (pk PairKeeper) transferToken(ctx sdk.Context, from, to sdk.AccAddress, token string, amount sdk.Int) {
	coin := newCoins(token, amount)
	if err := pk.UnFreezeCoins(ctx, from, coin); err != nil {
		panic(err)
	}
	if err := pk.SendCoins(ctx, from, to, coin); err != nil {
		panic(err)
	}
}

func newCoins(token string, amount sdk.Int) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(token, amount))
}

func (pk PairKeeper) HasOrder(ctx sdk.Context, orderID string) bool {
	return pk.GetOrder(ctx, orderID) != nil
}

func (pk PairKeeper) GetOrder(ctx sdk.Context, orderID string) *types.Order {
	return pk.IOrderBookKeeper.GetOrder(ctx, &QueryOrderInfo{
		OrderID: orderID,
	})
}

func (pk *PairKeeper) DeleteOrder(ctx sdk.Context, msg *types.MsgCancelOrder) sdk.Error {
	order := pk.GetOrder(ctx, msg.OrderID)
	if order == nil {
		return types.ErrInvalidOrderID(msg.OrderID)
	}
	if !order.Sender.Equals(msg.Sender) {
		return types.ErrInvalidSender(msg.Sender, order.Sender)
	}
	pk.IOrderBookKeeper.DelOrder(ctx, order)
	return pk.updateOrderBookReserveByOrderDel(ctx, order)
}

func (pk PairKeeper) updateOrderBookReserveByOrderDel(ctx sdk.Context, delOrder *types.Order) sdk.Error {
	info := pk.GetPoolInfo(ctx, delOrder.TradingPair)
	if delOrder.IsBuy {
		amount := delOrder.ActualAmount()
		info.MoneyOrderBookReserve = info.MoneyOrderBookReserve.Sub(amount)
		if err := pk.UnFreezeCoins(ctx, delOrder.Sender, newCoins(delOrder.Money(), amount)); err != nil {
			return err
		}
	} else {
		info.StockOrderBookReserve = info.StockOrderBookReserve.Sub(sdk.NewInt(delOrder.LeftStock))
		if err := pk.UnFreezeCoins(ctx, delOrder.Sender, newCoins(delOrder.Stock(), sdk.NewInt(delOrder.LeftStock))); err != nil {
			return err
		}
	}
	pk.SetPoolInfo(ctx, delOrder.TradingPair, info)
	return nil
}
