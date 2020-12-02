package autoswap

import (
	"github.com/coinexchain/cet-sdk/modules/incentive"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, keeper *Keeper) {
	keeper.ResetOrderIndexInOneBlock()
}

func BeginBlocker(ctx sdk.Context, keeper *Keeper) {
	if ctx.BlockHeight() == incentive.Dex3StartHeight {
		keeper.SetParams(ctx, DefaultParams())
	}
}
