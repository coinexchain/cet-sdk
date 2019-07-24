package asset

import (
	"os"
	"testing"

	"github.com/coinexchain/dex/types"
	dex "github.com/coinexchain/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	dex.InitSdkConfig()
	os.Exit(m.Run())
}

func TestGenesis(t *testing.T) {
	input := createTestInput()
	owner, _ := sdk.AccAddressFromBech32("coinex15fvnexrvsm9ryw3nn4mcrnqyhvhazkkrd4aqvd")
	input.tk.SetParams(input.ctx, DefaultParams())

	state := DefaultGenesisState()

	cet := &BaseToken{
		Name:             "CoinEx Chain Native Token",
		Symbol:           "cet",
		TotalSupply:      sdk.NewInt(588788547005740000),
		Owner:            owner,
		Mintable:         false,
		Burnable:         true,
		AddrForbiddable:  true,
		TokenForbiddable: true,
		TotalBurn:        sdk.NewInt(411211452994260000),
		TotalMint:        sdk.ZeroInt(),
		IsForbidden:      false,
	}
	abc := &BaseToken{
		Name:             "ABC Chain Native Token",
		Symbol:           "abc",
		TotalSupply:      sdk.NewInt(588788547005740000),
		Owner:            owner,
		Mintable:         false,
		Burnable:         true,
		AddrForbiddable:  true,
		TokenForbiddable: true,
		TotalBurn:        sdk.NewInt(411211452994260000),
		TotalMint:        sdk.ZeroInt(),
		IsForbidden:      false,
	}
	abcDump := &BaseToken{
		Name:             "ABC Chain Native Token",
		Symbol:           "abc",
		TotalSupply:      sdk.NewInt(588788547005740000),
		Owner:            owner,
		Mintable:         false,
		Burnable:         true,
		AddrForbiddable:  true,
		TokenForbiddable: true,
		TotalBurn:        sdk.NewInt(411211452994260000),
		TotalMint:        sdk.ZeroInt(),
		IsForbidden:      false,
	}
	abcInvalid := &BaseToken{
		Name:             "ABC Chain Native Token",
		Symbol:           "933",
		TotalSupply:      sdk.NewInt(588788547005740000),
		Owner:            owner,
		Mintable:         false,
		Burnable:         true,
		AddrForbiddable:  true,
		TokenForbiddable: true,
		TotalBurn:        sdk.NewInt(411211452994260000),
		TotalMint:        sdk.ZeroInt(),
		IsForbidden:      false,
	}
	state.Tokens = append(state.Tokens, cet, abc, abcDump, abcInvalid)
	require.Error(t, ValidateGenesis(state))
	state.Tokens = state.Tokens[:2]

	whitelist := []string{"cet:coinex1y5kdxnzn2tfwayyntf2n28q8q2s80mcul852ke"}
	state.Whitelist = append(state.Whitelist, whitelist...)

	forbiddenList := []string{"abc:coinex1p9ek7d3r9z4l288v4lrkwwrnh9k5htezk2q68g"}
	state.ForbiddenAddresses = append(state.ForbiddenAddresses, forbiddenList...)

	require.NoError(t, ValidateGenesis(state))
	InitGenesis(input.ctx, input.tk, state)

	res := input.tk.GetWhitelist(input.ctx, "cet")
	require.Equal(t, 1, len(res))
	require.Equal(t, "coinex1y5kdxnzn2tfwayyntf2n28q8q2s80mcul852ke", res[0].String())

	res = input.tk.GetForbiddenAddresses(input.ctx, "abc")
	require.Equal(t, 1, len(res))
	require.Equal(t, "coinex1p9ek7d3r9z4l288v4lrkwwrnh9k5htezk2q68g", res[0].String())

	export := ExportGenesis(input.ctx, input.tk)
	require.Equal(t, types.NewCetCoins(IssueTokenFee), export.Params.IssueTokenFee)
	require.Equal(t, types.NewCetCoins(IssueRareTokenFee), export.Params.IssueRareTokenFee)
	require.Equal(t, 2, len(export.Tokens))
	require.Equal(t, whitelist, export.Whitelist)
	require.Equal(t, forbiddenList, export.ForbiddenAddresses)
}
