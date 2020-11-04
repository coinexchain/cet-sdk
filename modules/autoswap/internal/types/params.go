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
	TakerFeeRateRate    int64
	MakerFeeRateRate    int64
	DealWithPoolFeeRate int64
	FeeToPool           int64
	FeeToValidator      int64
}

func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParams() Params {
	return Params{
		TakerFeeRateRate:    DefaultTakerFeeRate,
		MakerFeeRateRate:    DefaultMakerFeeRate,
		DealWithPoolFeeRate: DefaultDealWithPoolFeeRate,
		FeeToPool:           DefaultFeeToPool,
		FeeToValidator:      DefaultFeeToValidator,
	}
}

func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: keyTakerFeeRateRate, Value: &p.TakerFeeRateRate},
		{Key: keyMakerFeeRate, Value: &p.MakerFeeRateRate},
		{Key: keyDealWithPoolFeeRate, Value: &p.DealWithPoolFeeRate},
		{Key: keyFeeToPool, Value: &p.FeeToPool},
		{Key: keyFeeToValidator, Value: &p.FeeToValidator},
	}
}

func (p *Params) ValidateGenesis() error {
	if p.TakerFeeRateRate <= 0 || p.MakerFeeRateRate <= 0 || p.DealWithPoolFeeRate <= 0 || p.FeeToPool <= 0 || p.FeeToValidator <= 0 {
		return fmt.Errorf("parameter can not be a negative number: TakerFeeRate: %d, "+
			"MakerFeeRate: %d, DealWithPoolFeeRate: %d, FeeToPool: %d, FeeToValidator: %d", p.TakerFeeRateRate, p.MakerFeeRateRate, p.DealWithPoolFeeRate, p.FeeToPool, p.FeeToValidator)
	}
	if p.TakerFeeRateRate >= DefaultFeePrecision || p.MakerFeeRateRate >= DefaultFeePrecision || p.DealWithPoolFeeRate >= DefaultFeePrecision {
		return fmt.Errorf("FeeRate should be less than 1. TakerFeeRate: %d, MakerFeeRate: %d, DealWithPoolFeeRate: %d",
			p.TakerFeeRateRate, p.MakerFeeRateRate, p.DealWithPoolFeeRate)
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
		p.TakerFeeRateRate,
		p.MakerFeeRateRate,
		p.DealWithPoolFeeRate,
		p.FeeToPool,
		p.FeeToValidator)
}
