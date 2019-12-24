package comment

import (
	"github.com/coinexchain/cet-sdk/modules/comment/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/comment/internal/types"
)

const (
	StoreKey   = types.StoreKey
	ModuleName = types.ModuleName
)

var (
	NewBaseKeeper = keepers.NewKeeper
)

type (
	Keeper          = keepers.Keeper
	TokenComment    = types.TokenComment
	CommentRef      = types.CommentRef
	MsgCommentToken = types.MsgCommentToken
)
