package types

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestOrderCommission(t *testing.T) {
	bigInt, ok := sdk.NewIntFromString("57896044618658097711785492504343953926634992332820282019728792003956564819967")
	require.Equal(t, true, ok)
	bigDec := bigInt.ToDec()
	require.Equal(t, 40, len(bigDec.Int.Bytes()))
	maxDec := sdk.NewDec(1)
	for i := 0; i < 255; i++ {
		maxDec = maxDec.MulInt64(2)
		fmt.Printf("%v %d \n", maxDec.Int.Bytes(), maxDec)
	}
	require.Equal(t, 40, len(maxDec.Int.Bytes()))

	addr, failed := sdk.AccAddressFromHex("0123456789012345678901234567890123423456")
	require.Nil(t, failed)
	order := Order{
		Sender:   addr,
		Sequence: 9223372036854775818,
		Identify: 28,
	}
	require.Equal(t, addr.String()+"-2361183241434822609436", order.OrderID())
	order = Order{
		Sender:   addr,
		Sequence: 9223372036854775829,
		Identify: 255,
	}
	require.Equal(t, addr.String()+"-2361183241434822612479", order.OrderID())

	bz1 := DecToBigEndianBytes(sdk.NewDec(math.MaxInt64).MulInt64(100))
	bz2 := DecToBigEndianBytes(sdk.NewDec(-math.MaxInt64).MulInt64(100))
	require.Equal(t, bz1, bz2)
	bz2 = DecToBigEndianBytes(sdk.NewDec(math.MaxInt64).MulInt64(100).Add(sdk.NewDec(1)))
	require.Equal(t, 1, bytes.Compare(bz2, bz1))

	order.DealStock = 0
	order.FrozenCommission = 10000
	order.Quantity = 100000
	require.Equal(t, int64(100), order.CalActualOrderCommissionInt64(100))
	order.DealStock = 50000
	require.Equal(t, int64(5000), order.CalActualOrderCommissionInt64(100))
	order.DealStock = 50009
	require.Equal(t, int64(5000), order.CalActualOrderCommissionInt64(100))
	order.DealStock = 50010
	require.Equal(t, int64(5001), order.CalActualOrderCommissionInt64(100))
	order.FrozenCommission = MaxOrderAmount + 10
	order.DealStock = 100000
	require.Equal(t, MaxOrderAmount, order.CalActualOrderCommissionInt64(100))
}

func TestOrder_CalActualOrderFeatureFeeInt64(t *testing.T) {
	addr, failed := sdk.AccAddressFromHex("0123456789012345678901234567890123423456")
	require.Nil(t, failed)
	order := Order{
		Sender:           addr,
		Sequence:         9223,
		Identify:         28,
		Height:           9,
		FrozenFeatureFee: 100,
		ExistBlocks:      200,
	}

	ctx := sdk.Context{}
	ctx = ctx.WithBlockHeight(10)
	fee := order.CalActualOrderFeatureFeeInt64(ctx, 100)
	require.EqualValues(t, 0, fee)

	ctx = ctx.WithBlockHeight(109)
	fee = order.CalActualOrderFeatureFeeInt64(ctx, 100)
	require.EqualValues(t, 1, fee)

	ctx = ctx.WithBlockHeight(200)
	fee = order.CalActualOrderFeatureFeeInt64(ctx, 100)
	require.EqualValues(t, 92, fee)

	ctx = ctx.WithBlockHeight(208)
	fee = order.CalActualOrderFeatureFeeInt64(ctx, 100)
	require.EqualValues(t, 100, fee)

	ctx = ctx.WithBlockHeight(20800)
	fee = order.CalActualOrderFeatureFeeInt64(ctx, 100)
	require.EqualValues(t, 100, fee)

	ctx = ctx.WithBlockHeight(210)
	fee = order.CalActualOrderFeatureFeeInt64(ctx, 200)
	require.EqualValues(t, 0, fee)
}

func TestFuzz_CalActualOrderCommissionInt64(t *testing.T) {
	param := DefaultParams()
	r := rand.New(rand.NewSource(2))
	addr, _ := sdk.AccAddressFromHex("0123456789012345678901234567890123423456")
	for i := 0; i < 500; i++ {
		or := getRandOrder(r, addr, &param)
		or.CalActualOrderCommissionInt64(param.FeeForZeroDeal)
	}
}

func TestFuzz_CalActualOrderFeatureFeeInt64(t *testing.T) {
	ctx := sdk.Context{}
	param := DefaultParams()
	r := rand.New(rand.NewSource(2))
	addr, _ := sdk.AccAddressFromHex("0123456789012345678901234567890123423456")

	for i := 0; i < 500; i++ {
		ctx = ctx.WithBlockHeight(r.Int63n(10000))
		or := getRandOrder(r, addr, &param)
		or.CalActualOrderFeatureFeeInt64(ctx, param.GTEOrderLifetime)
	}
}

func getRandOrder(r *rand.Rand, addr sdk.AccAddress, param *Params) *Order {
	side := SELL
	if r.Int31n(2) == 1 {
		side = BUY
	}
	tif := GTE
	if r.Int31n(4)%3 == 0 {
		tif = IOC
	}
	or := Order{
		Sender:           addr,
		Sequence:         r.Uint64(),
		Identify:         0,
		TradingPair:      "abc/cet",
		OrderType:        LimitOrder,
		Price:            sdk.NewDec(r.Int63n(1000000)),
		Quantity:         r.Int63n(100000),
		Side:             byte(side),
		TimeInForce:      int64(tif),
		Height:           r.Int63n(10000),
		FrozenCommission: r.Int63n(param.FeeForZeroDeal * 1000),
		ExistBlocks:      r.Int63n(param.GTEOrderLifetime * 100),
		FrozenFeatureFee: r.Int63n(param.FeeForZeroDeal * 100),
		FrozenFee:        r.Int63n(10000),
		LeftStock:        0,
		Freeze:           0,
		DealStock:        r.Int63n(10000),
		DealMoney:        r.Int63n(10000),
	}
	return &or
}
