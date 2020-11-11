package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

var _ params.ParamSet = (*Params)(nil)

var (
	KeyIncentiveDefaultRewardPerBlock = []byte("incentiveDefaultRewardPerBlock")
	KeyIncentivePlans                 = []byte("incentivePlans")
	KeyIncentiveUpdatedRewards        = []byte("updatedRewards")
)

type Params struct {
	DefaultRewardPerBlock int64  `json:"default_reward_per_block"`
	Plans                 []Plan `json:"plans"`
}

type Plan struct {
	StartHeight    int64 `json:"start_height"`
	EndHeight      int64 `json:"end_height"`
	RewardPerBlock int64 `json:"reward_per_block"`
	TotalIncentive int64 `json:"total_incentive"`
}

func DefaultParams() Params {
	return Params{
		DefaultRewardPerBlock: 3e8,
		Plans:                 []Plan{},
	}
}

// ParamKeyTable type declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyIncentiveDefaultRewardPerBlock, Value: &p.DefaultRewardPerBlock},
		{Key: KeyIncentivePlans, Value: &p.Plans},
	}
}

func (p Params) String() string {
	s := fmt.Sprintf(`Incentive Params:
  DefaultRewardPerBlock: %d`,
		p.DefaultRewardPerBlock)

	for _, p := range p.Plans {
		s += fmt.Sprintf("\n  Plan: StartHeight=%d EndHeight=%d RewardPerBlock=%d TotalIncentive=%d",
			p.StartHeight, p.EndHeight, p.RewardPerBlock, p.TotalIncentive)
	}

	return s
}

func CheckPlans(plans []Plan) sdk.Error {

	for _, plan := range plans {
		if plan.StartHeight < 0 || plan.EndHeight < 0 {
			return sdk.NewError(CodeSpaceIncentive, CodeInvalidPlanHeight, "invalid incentive plan height")
		}
		if plan.EndHeight <= plan.StartHeight {
			return sdk.NewError(CodeSpaceIncentive, CodeInvalidPlanHeight, "incentive plan end height should be greater than start height")
		}
		if plan.RewardPerBlock < 0 {
			return sdk.NewError(CodeSpaceIncentive, CodeInvalidRewardPerBlock, "invalid incentive plan reward per block")
		}
		if plan.TotalIncentive < 0 {
			return sdk.NewError(CodeSpaceIncentive, CodeInvalidTotalIncentive, "invalid incentive plan total incentive reward")
		}
		if (plan.EndHeight-plan.StartHeight)*plan.RewardPerBlock != plan.TotalIncentive {
			return sdk.NewError(CodeSpaceIncentive, CodeInvalidTotalIncentive, "invalid incentive plan")
		}
	}
	return nil
}
