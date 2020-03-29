package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestParams_Equal(t *testing.T) {
	param := DefaultParams()
	param2 := NewParams(sdk.MustNewDecFromStr("20.0"), 7*24*60*60*1000000000, 2000)
	b := param.Equal(param2)
	require.Equal(t, true, b)
}
