package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestMaxAmount(t *testing.T) {
	require.EqualValues(t, true, !MaxAmount.IsZero())
}

func TestDec(t *testing.T) {
	p, _ := sdk.NewDecFromStr("1.023")
	fmt.Println(p.String())
	fmt.Println(p.TruncateInt().String())
	fmt.Println(p.TruncateDec().String())

	fmt.Println(sdk.NewInt(-1))
}
