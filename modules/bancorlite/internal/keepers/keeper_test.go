package keepers_test

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/authx"
	"github.com/coinexchain/cet-sdk/modules/bancorlite/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/bancorlite/internal/types"
	"github.com/coinexchain/cet-sdk/testapp"
)

var (
	owner = sdk.AccAddress("user")
	bch   = "bch"
	cet   = "cet"
	abc   = "abc"
)

func defaultContext() (keepers.Keeper, sdk.Context) {
	app := testapp.NewTestApp()
	ctx := sdk.NewContext(app.Cms, abci.Header{}, false, log.NewNopLogger())
	return app.BancorKeeper, ctx
}
func TestBancorInfo_UpdateStockInPool(t *testing.T) {
	type fields struct {
		Owner              sdk.AccAddress
		Stock              string
		Money              string
		InitPrice          sdk.Dec
		MaxSupply          sdk.Int
		MaxPrice           sdk.Dec
		Price              sdk.Dec
		StockInPool        sdk.Int
		MoneyInPool        sdk.Int
		EarliestCancelTime int64
	}
	type args struct {
		stockInPool sdk.Int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "positive",
			fields: fields{
				Owner:              owner,
				Stock:              bch,
				Money:              cet,
				InitPrice:          sdk.NewDec(0),
				MaxSupply:          sdk.NewInt(100),
				MaxPrice:           sdk.NewDec(10),
				StockInPool:        sdk.NewInt(10),
				EarliestCancelTime: 100,
			},
			args: args{
				stockInPool: sdk.NewInt(20),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bi := &keepers.BancorInfo{
				Owner:              tt.fields.Owner,
				Stock:              tt.fields.Stock,
				Money:              tt.fields.Money,
				InitPrice:          tt.fields.InitPrice,
				MaxSupply:          tt.fields.MaxSupply,
				MaxPrice:           tt.fields.MaxPrice,
				MaxMoney:           sdk.ZeroInt(),
				Price:              tt.fields.Price,
				StockInPool:        tt.fields.StockInPool,
				MoneyInPool:        tt.fields.MoneyInPool,
				AR:                 0,
				EarliestCancelTime: tt.fields.EarliestCancelTime,
			}
			if got := bi.UpdateStockInPool(tt.args.stockInPool); got != tt.want {
				t.Errorf("BancorInfo.UpdateStockInPool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBancorInfo_UpdateStockInPool2(t *testing.T) {

	rand.Seed(time.Now().Unix())
	random := func(min, max int64) int64 {
		return rand.Int63n(max-min) + min
	}
	maxSupply := int64(10000000000000)
	times := 1000000
	count := 0
	count2 := 0
	count3 := 0
	for i := 0; i < times; i++ {
		initPrice := random(0, 10)
		maxPrice := random(10, 20)
		randomAr := random(0, 5000)
		var tmpAr float64
		tmpAr = float64(randomAr) / 1000
		if randomAr == 0 {
			tmpAr = float64(randomAr) + 0.1
		}
		maxMoney := int64((float64(maxSupply*maxPrice) + tmpAr*float64(initPrice)*float64(maxSupply)) / (tmpAr + 1))
		supply := random(0, maxSupply)

		bi := keepers.BancorInfo{
			Owner:              nil,
			Stock:              "",
			Money:              "",
			InitPrice:          sdk.NewDec(initPrice),
			MaxSupply:          sdk.NewInt(maxSupply),
			StockPrecision:     0,
			MaxPrice:           sdk.NewDec(maxPrice),
			MaxMoney:           sdk.NewInt(maxMoney),
			AR:                 0,
			Price:              sdk.NewDec(initPrice),
			StockInPool:        sdk.NewInt(maxSupply),
			MoneyInPool:        sdk.NewInt(0),
			EarliestCancelTime: 0,
		}

		ar, money, _ := CalculateMoney(float64(supply), float64(maxSupply), float64(maxMoney), float64(initPrice), float64(maxPrice))
		bi.AR = int64(ar * types.ARSamples)
		bi.UpdateStockInPool(sdk.NewInt(maxSupply - supply))
		diffMoney := math.Abs(money - float64(bi.MoneyInPool.Int64()))
		if diffMoney/money > 0.000001 {
			//fmt.Printf("money is rough: ar:%f, AR in pool:%d, money diff ratio:%f, maxMoney:%d, maxSupply:%d, initPrice:%d, maxPrice:%d," +
			//	" supply:%d, money:%f, moneyInPool:%d, price:%f, priceInPool:%s" +
			//	"\n",
			//	ar, bi.AR, diffMoney/money, maxMoney, maxSupply, initPrice, maxPrice, supply, money, bi.MoneyInPool.Int64(),
			//	price, bi.Price.String(), )
			count++
		}
		if diffMoney/money > 0.00001 {
			count2++
		}
		if diffMoney/money > 0.0001 {
			count3++
		}
		//s := fmt.Sprintf("%f", price)
		//priceDec, _ := sdk.NewDecFromStr(s)
		//fmt.Printf("priceDec:%s\n", priceDec.String())
		//if priceDec.Sub(bi.Price).GT(sdk.NewDec(1)) || bi.Price.Sub(priceDec).GT(sdk.NewDec(1)) {
		//	fmt.Printf("price is rough: ar:%d, maxMoney:%d, maxSupply:%d, initPrice:%d, maxPrice:%d, supply:%d, price:%f, priceInPool:%s\n",
		//		bi.AR, maxMoney, maxSupply, initPrice, maxPrice, supply, price, bi.Price.String())
		//}
	}
	//fmt.Printf("percent of pass 0.000_001: %f\n", 1 - float64(count)/float64(times))
	//fmt.Printf("percent of pass 0.000_01: %f\n", 1 - float64(count2)/float64(times))
	//fmt.Printf("percent of pass 0.000_1: %f\n", 1 - float64(count3)/float64(times))
	require.True(t, float64(count3)/float64(times) < 0.05)
}

func CalculateMoney(supply, maxSupply, maxMoney, initPrice, maxPrice float64) (ar, money, price float64) {
	ar = (maxSupply*maxPrice - maxMoney) / (maxMoney - initPrice*maxSupply)
	money = math.Pow(supply/maxSupply, ar+1)*(maxMoney-initPrice*maxSupply) + initPrice*supply
	price = math.Pow(supply/maxSupply, ar)*(maxPrice-initPrice) + initPrice
	return
}

func TestBancorInfo_IsConsistent(t *testing.T) {
	type fields struct {
		Owner              sdk.AccAddress
		Stock              string
		Money              string
		InitPrice          sdk.Dec
		MaxSupply          sdk.Int
		MaxPrice           sdk.Dec
		Price              sdk.Dec
		StockInPool        sdk.Int
		MoneyInPool        sdk.Int
		EarliestCancelTime int64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "positive",
			fields: fields{
				Owner:              owner,
				Stock:              bch,
				Money:              cet,
				InitPrice:          sdk.NewDec(0),
				MaxSupply:          sdk.NewInt(100),
				MaxPrice:           sdk.NewDec(10),
				Price:              sdk.NewDec(1),
				StockInPool:        sdk.NewInt(90),
				MoneyInPool:        sdk.NewInt(5),
				EarliestCancelTime: 100,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bi := &keepers.BancorInfo{
				Owner:              tt.fields.Owner,
				Stock:              tt.fields.Stock,
				Money:              tt.fields.Money,
				InitPrice:          tt.fields.InitPrice,
				MaxSupply:          tt.fields.MaxSupply,
				MaxPrice:           tt.fields.MaxPrice,
				MaxMoney:           sdk.ZeroInt(),
				Price:              tt.fields.Price,
				StockInPool:        tt.fields.StockInPool,
				MoneyInPool:        tt.fields.MoneyInPool,
				EarliestCancelTime: tt.fields.EarliestCancelTime,
			}
			if got := bi.IsConsistent(); got != tt.want {
				t.Errorf("BancorInfo.IsConsistent() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestBancorInfoKeeper(t *testing.T) {
	keeper, ctx := defaultContext()
	bi := []keepers.BancorInfo{
		{
			Owner:              owner,
			Stock:              bch,
			Money:              cet,
			InitPrice:          sdk.NewDec(0),
			MaxSupply:          sdk.NewInt(100),
			MaxPrice:           sdk.NewDec(10),
			MaxMoney:           sdk.ZeroInt(),
			AR:                 1,
			Price:              sdk.NewDec(1),
			StockInPool:        sdk.NewInt(90),
			MoneyInPool:        sdk.NewInt(5),
			EarliestCancelTime: 100,
		},
		{
			Owner:              owner,
			Stock:              abc,
			Money:              cet,
			InitPrice:          sdk.NewDec(0),
			MaxSupply:          sdk.NewInt(100),
			MaxPrice:           sdk.NewDec(10),
			MaxMoney:           sdk.ZeroInt(),
			AR:                 1,
			Price:              sdk.NewDec(1),
			StockInPool:        sdk.NewInt(90),
			MoneyInPool:        sdk.NewInt(5),
			EarliestCancelTime: 0,
		},
	}
	for _, p := range bi {
		keeper.Save(ctx, &p)
	}

	for i := range bi {
		loadBI := keeper.Load(ctx, bi[i].GetSymbol())
		require.True(t, reflect.DeepEqual(*loadBI, bi[i]))
	}

	keeper.Remove(ctx, &bi[0])
	require.Nil(t, keeper.Load(ctx, bi[0].GetSymbol()))
	require.False(t, keeper.IsBancorExist(ctx, bch))
	require.True(t, keeper.IsBancorExist(ctx, abc))
}

func TestKeeper_GetRebate(t *testing.T) {
	app := testapp.NewTestApp()
	referee := sdk.AccAddress("referee")
	ctx := sdk.NewContext(app.Cms, abci.Header{}, false, log.NewNopLogger())
	app.AccountXKeeper.SetParams(ctx, authx.DefaultParams())
	app.AccountXKeeper.SetAccountX(ctx, authx.NewAccountX(owner, false, nil, nil, referee, 0))
	acc, rebate, balance, exist := app.BancorKeeper.GetRebate(ctx, owner, sdk.NewInt(100000))
	require.Equal(t, acc, referee)
	require.Equal(t, int64(20000), rebate.Int64())
	require.Equal(t, int64(80000), balance.Int64())
	require.Equal(t, exist, true)
}

func TestCurrentPriceCalculate(t *testing.T) {
	bi := keepers.BancorInfo{
		Owner:              nil,
		Stock:              "btc",
		Money:              "cet",
		InitPrice:          sdk.NewDec(3),
		MaxSupply:          sdk.NewInt(10000000000000),
		StockPrecision:     6,
		MaxPrice:           sdk.NewDec(10),
		MaxMoney:           sdk.NewInt(70000000000000),
		AR:                 750,
		Price:              sdk.NewDec(3),
		StockInPool:        sdk.NewInt(100000_0000_0000),
		MoneyInPool:        sdk.NewInt(0),
		EarliestCancelTime: 0,
	}

	biNew := bi
	display := keepers.NewBancorInfoDisplay(&biNew)
	require.Equal(t, display.CurrentPrice, bi.InitPrice.String())
	biNew.UpdateStockInPool(biNew.StockInPool.Sub(sdk.NewInt(1_0000_0000)))
	display = keepers.NewBancorInfoDisplay(&biNew)
	biNew.UpdateStockInPool(biNew.StockInPool.Sub(sdk.NewInt(99_0000_0000)))

	priceRatio := types.TableLookup(1750, 1)
	pp := biNew.InitPrice.MulInt(sdk.NewInt(100_0000_0000)).Add(biNew.MaxPrice.Sub(biNew.InitPrice).MulInt(biNew.MaxSupply).QuoInt64(175).MulInt64(100).Mul(priceRatio)).QuoInt64(100_0000_0000)
	for i := 0; i < 10; i++ {
		require.Equal(t, display.CurrentPrice[i], pp.String()[i])
	}
}

func TestPrice(t *testing.T) {
	p, _ := sdk.NewDecFromStr("2.013267155742863000")
	bi := keepers.BancorInfo{
		Owner:              nil,
		Stock:              "btc",
		Money:              "cet",
		InitPrice:          sdk.NewDec(2),
		MaxSupply:          sdk.NewInt(200_0000_0000),
		StockPrecision:     4,
		MaxPrice:           sdk.NewDec(5),
		MaxMoney:           sdk.NewInt(600_0000_0000),
		AR:                 2000,
		Price:              p,
		StockInPool:        sdk.NewInt(186_7000_0000),
		MoneyInPool:        sdk.NewInt(2665902715 - 20133),
		EarliestCancelTime: 0,
	}
	//for i := 0; i < 9; i++ {
	biNew := bi
	stockInPool := bi.StockInPool.SubRaw(1_0000)
	biNew.UpdateStockInPool(stockInPool)
	diff := sdk.NewDecFromInt(biNew.MoneyInPool.Sub(bi.MoneyInPool)).QuoInt64(1_0000)
	//fmt.Println(diff, "cet")
	bi = biNew
	biNew = bi
	stockInPool = bi.StockInPool.SubRaw(1_0000)
	biNew.UpdateStockInPool(stockInPool)
	diff = sdk.NewDecFromInt(biNew.MoneyInPool.Sub(bi.MoneyInPool)).QuoInt64(1_0000)
	//fmt.Println(diff, "cet")
	bi = biNew
	biNew = bi
	stockInPool = bi.StockInPool.SubRaw(1_0000)
	biNew.UpdateStockInPool(stockInPool)
	diff = sdk.NewDecFromInt(biNew.MoneyInPool.Sub(bi.MoneyInPool)).QuoInt64(1_0000)
	//fmt.Println(diff, "cet")
	bi = biNew
	biNew = bi
	stockInPool = bi.StockInPool.SubRaw(100_0000)
	biNew.UpdateStockInPool(stockInPool)
	diff = sdk.NewDecFromInt(biNew.MoneyInPool.Sub(bi.MoneyInPool)).QuoInt64(100_0000)
	fmt.Println(diff, "cet")
	bi = biNew
	//}
}
