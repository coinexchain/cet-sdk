package types

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	DefaultFeePrecision        = 10000
	DefaultTakerFeeRate        = 30
	DefaultMakerFeeRate        = 50
	DefaultDealWithPoolFeeRate = 50
	DefaultFeeToPool           = 6
	DefaultFeeToValidator      = 4
)

var (
	keyTakerFeeRateRate    = []byte("TakerFeeRate")
	keyMakerFeeRate        = []byte("MakerFeeRate")
	keyDealWithPoolFeeRate = []byte("DealWithPoolFeeRate")
	keyFeeToPool           = []byte("FeeToPool")
	keyFeeToValidator      = []byte("FeeToValidator")
)

type Params struct {
	TakerFeeRate        int64 `json:"taker_fee_rate"`
	MakerFeeRate        int64 `json:"maker_fee_rate"`
	DealWithPoolFeeRate int64 `json:"deal_with_pool_fee_rate"`
	FeeToPool           int64 `json:"fee_to_pool"`
	FeeToValidator      int64 `json:"fee_to_validator"`
}

func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParams() Params {
	return Params{
		TakerFeeRate:        DefaultTakerFeeRate,
		MakerFeeRate:        DefaultMakerFeeRate,
		DealWithPoolFeeRate: DefaultDealWithPoolFeeRate,
		FeeToPool:           DefaultFeeToPool,
		FeeToValidator:      DefaultFeeToValidator,
	}
}

func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: keyTakerFeeRateRate, Value: &p.TakerFeeRate},
		{Key: keyMakerFeeRate, Value: &p.MakerFeeRate},
		{Key: keyDealWithPoolFeeRate, Value: &p.DealWithPoolFeeRate},
		{Key: keyFeeToPool, Value: &p.FeeToPool},
		{Key: keyFeeToValidator, Value: &p.FeeToValidator},
	}
}

func (p *Params) ValidateGenesis() error {
	if p.TakerFeeRate <= 0 || p.MakerFeeRate <= 0 || p.DealWithPoolFeeRate <= 0 || p.FeeToPool <= 0 || p.FeeToValidator <= 0 {
		return fmt.Errorf("parameter can not be a negative number: TakerFeeRate: %d, "+
			"MakerFeeRate: %d, DealWithPoolFeeRate: %d, FeeToPool: %d, FeeToValidator: %d", p.TakerFeeRate, p.MakerFeeRate, p.DealWithPoolFeeRate, p.FeeToPool, p.FeeToValidator)
	}
	if p.TakerFeeRate >= DefaultFeePrecision || p.MakerFeeRate >= DefaultFeePrecision || p.DealWithPoolFeeRate >= DefaultFeePrecision {
		return fmt.Errorf("FeeRate should be less than 1. TakerFeeRate: %d, MakerFeeRate: %d, DealWithPoolFeeRate: %d",
			p.TakerFeeRate, p.MakerFeeRate, p.DealWithPoolFeeRate)
	}
	return nil
}

func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

func (p Params) String() string {
	return fmt.Sprintf(`autoswap Params:
	TakerFeeRate: %d,
	MakerFeeRate: %d,
	DealWithPoolFeeRate: %d,
	FeeToPool: %d,
	FeeToValidator: %d`,
		p.TakerFeeRate,
		p.MakerFeeRate,
		p.DealWithPoolFeeRate,
		p.FeeToPool,
		p.FeeToValidator)
}
