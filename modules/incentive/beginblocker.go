package incentive

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/coinexchain/cet-sdk/modules/incentive/internal/keepers"
	dex "github.com/coinexchain/cet-sdk/types"
)

var (
	Dex3StartHeight = int64(100000000) // TODO
)

func BeginBlocker(ctx sdk.Context, k keepers.Keeper) {
	if ctx.BlockHeight() >= Dex3StartHeight {
		collectRewardsFromPool(ctx, k)
	}
}

func collectRewardsFromPool(ctx sdk.Context, k Keeper) {
	rewardAmount := k.GetParams(ctx).DefaultRewardPerBlock
	blockRewards := sdk.NewCoins(sdk.NewInt64Coin(dex.DefaultBondDenom, rewardAmount))

	err := k.MintTokenByModule(ctx, dex.DefaultBondDenom, sdk.NewInt(rewardAmount), ModuleName)
	if err != nil {
		panic(err)
	}

	err = k.SendCoinsFromModuleToModule(ctx, ModuleName, auth.FeeCollectorName, blockRewards)
	if err != nil {
		panic(err)
	}

}
