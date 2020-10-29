package cli

import (
	"os"
	"testing"

	"github.com/coinexchain/cosmos-utils/client/cliutil"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var resultMsg cliutil.MsgWithAccAddress
var fromAddr sdk.AccAddress

var resultParam interface{}
var resultPath string

func cliRunCmdForTest(cdc *codec.Codec, msg cliutil.MsgWithAccAddress) error {
	cliCtx := context.NewCLIContext().WithCodec(cdc)
	senderAddr := cliCtx.GetFromAddress()
	msg.SetAccAddress(senderAddr)
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	resultMsg = msg
	return nil
}

func cliQueryForTest(cdc *codec.Codec, path string, param interface{}) error {
	resultParam = param
	resultPath = path
	return nil
}

func setup() {
	sdk.GetConfig().SetBech32PrefixForAccount("coinex", "coinexpub")
	fromAddr, _ = sdk.AccAddressFromHex("01234567890123456789012345678901234abcde")

	cliutil.CliRunCommand = cliRunCmdForTest
	cliutil.CliQuery = cliQueryForTest
}

// https://stackoverflow.com/questions/23729790/how-can-i-do-test-setup-using-the-testing-package-in-go
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	//shutdown()
	os.Exit(code)
}
