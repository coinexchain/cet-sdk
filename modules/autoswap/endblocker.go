package autoswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, keeper *Keeper) {
	keeper.ResetOrderIndexInOneBlock()
}
