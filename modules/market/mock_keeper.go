package market

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinexchain/cet-sdk/modules/asset"
)

type mockKeeper struct {
	records []string
}

func (k *mockKeeper) GetRefereeAddr(ctx sdk.Context, accAddr sdk.AccAddress) sdk.AccAddress {
	addr, err := sdk.AccAddressFromHex("0123456789012345678901234567890123400012")
	if err != nil {
		panic("generate address failed")
	}
	return addr
}
func (k *mockKeeper) GetRebateRatio(ctx sdk.Context) int64 {
	return 100
}
func (k *mockKeeper) GetRebateRatioBase(ctx sdk.Context) int64 {
	return 10000
}
func (k *mockKeeper) SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	panic("implement me")
}
func (k *mockKeeper) DeductInt64CetFee(ctx sdk.Context, addr sdk.AccAddress, amt int64) sdk.Error {
	panic("implement me")
}
func (k *mockKeeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	panic("implement me")
}
func (k *mockKeeper) SendCoins(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) sdk.Error {
	k.records = append(k.records, fmt.Sprintf("send %s %s from %s to %s",
		amt[0].Amount.String(), amt[0].Denom, from.String(), to.String()))
	return nil
}
func (k *mockKeeper) FreezeCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error {
	panic("implement me")
}
func (k *mockKeeper) UnFreezeCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error {
	panic("implement me")
}
func (k *mockKeeper) SubtractFeeAndCollectFee(ctx sdk.Context, addr sdk.AccAddress, amt int64) sdk.Error {
	fee := fmt.Sprintf("addr : %s, fee : %d", addr, amt)
	k.records = append(k.records, fee)
	return nil
}
func (k *mockKeeper) cleanRecord() {
	k.records = make([]string, 0, 2)
}

type mocAssertStatusKeeper struct {
	forbiddenDenomList       []string
	globalForbiddenDenomList []string
	forbiddenAddrList        []sdk.AccAddress
}

func (k *mocAssertStatusKeeper) IsTokenForbidden(ctx sdk.Context, denom string) bool {
	for _, d := range k.globalForbiddenDenomList {
		if denom == d {
			return true
		}
	}
	return false
}
func (k *mocAssertStatusKeeper) IsTokenExists(ctx sdk.Context, denom string) bool {
	return true
}
func (k *mocAssertStatusKeeper) IsTokenIssuer(ctx sdk.Context, denom string, addr sdk.AccAddress) bool {
	return false
}
func (k *mocAssertStatusKeeper) IsForbiddenByTokenIssuer(ctx sdk.Context, denom string, addr sdk.AccAddress) bool {
	for i := 0; i < len(k.forbiddenDenomList); i++ {
		if denom == k.forbiddenDenomList[i] && bytes.Equal(addr, k.forbiddenAddrList[i]) {
			return true
		}
	}
	return false
}
func (k *mocAssertStatusKeeper) GetToken(ctx sdk.Context, symbol string) asset.Token {
	return nil
}

type mocBankxKeeper struct {
	records []string
}

func (k *mocBankxKeeper) SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return nil
}
func (k *mocBankxKeeper) DeductInt64CetFee(ctx sdk.Context, addr sdk.AccAddress, amt int64) sdk.Error {
	return nil
}
func (k *mocBankxKeeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	return true
}
func (k *mocBankxKeeper) SendCoins(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) sdk.Error {
	k.records = append(k.records, fmt.Sprintf("send %s %s from %s to %s",
		amt[0].Amount.String(), amt[0].Denom, from.String(), to.String()))
	return nil
}
func (k *mocBankxKeeper) FreezeCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error {
	k.records = append(k.records, fmt.Sprintf("freeze %s %s at %s",
		amt[0].Amount.String(), amt[0].Denom, string(acc)))
	return nil
}
func (k *mocBankxKeeper) UnFreezeCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) sdk.Error {
	k.records = append(k.records, fmt.Sprintf("unfreeze %s %s at %s",
		amt[0].Amount.String(), amt[0].Denom, acc.String()))
	return nil
}
