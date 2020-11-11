package incentive

import (
	"github.com/coinexchain/cet-sdk/modules/incentive/internal/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/coinexchain/cet-sdk/modules/incentive/internal/keepers"
	dex "github.com/coinexchain/cet-sdk/types"
)

var (
	PoolAddr        = sdk.AccAddress(crypto.AddressHash([]byte("incentive_pool")))
	Dex3StartHeight = int64(100000000) // TODO
)

func BeginBlocker(ctx sdk.Context, k keepers.Keeper) {
	if ctx.BlockHeight() == Dex3StartHeight {
		clearIncentiveState(ctx, k)
	}

	collectRewardsFromPool(ctx, k)
}

func collectRewardsFromPool(ctx sdk.Context, k Keeper) {
	rewardAmount := k.GetParams(ctx).DefaultRewardPerBlock
	blockRewards := sdk.NewCoins(sdk.NewInt64Coin(dex.DefaultBondDenom, rewardAmount))
	err := k.MintCoins(ctx, ModuleName, blockRewards)
	if err != nil {
		panic(err)
	}

	err = k.SendCoinsFromModuleToModule(ctx, ModuleName, auth.FeeCollectorName, blockRewards)
	if err != nil {
		panic(err)
	}
}
func clearIncentiveState(ctx sdk.Context, k Keeper) {
	// burn all coins in PoolAddr
	allCoins := k.GetCoins(ctx, PoolAddr)
	err := k.SendCoinsFromAccountToModule(ctx, PoolAddr, types.ModuleName, allCoins)
	if err != nil {
		panic(err)
	}
	err = k.BurnCoins(ctx, types.ModuleName, allCoins)
	if err != nil {
		panic(err)
	}

	// clear unused plans & params
	k.SetParams(ctx, types.DefaultParams())
}
