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
func clearIncentiveState(ctx sdk.Context, k Keeper) {
	// clear unused plans & params
	k.SetParams(ctx, types.DefaultParams())

	// burn pool's cet token
	allCoins := k.GetCoins(ctx, PoolAddr)
	for _, coin := range allCoins {
		if coin.Denom == dex.DefaultBondDenom {
			err := k.SendCoinsFromAccountToModule(ctx, PoolAddr, ModuleName, sdk.NewCoins(coin))
			if err != nil {
				panic(err)
			}
			err = k.BurnTokenByModule(ctx, coin.Denom, coin.Amount, ModuleName)
			if err != nil {
				panic(err)
			}
		}
	}
	// update cet to be mintable
	err := k.UpdateCETMintable(ctx)
	if err != nil {
		panic(err)
	}
}
