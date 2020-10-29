package rest

import (
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var addr sdk.AccAddress

func setup() {
	sdk.GetConfig().SetBech32PrefixForAccount("coinex", "coinexpub")
	addr, _ = sdk.AccAddressFromBech32("coinex1px8alypku5j84qlwzdpynhn4nyrkagaytu5u4a")
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	//shutdown()
	os.Exit(code)
}
