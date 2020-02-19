package authx

import (
	"github.com/coinexchain/cet-sdk/modules/authx/internal/keepers"
	"github.com/coinexchain/cet-sdk/modules/authx/internal/types"
)

const (
	StoreKey        = types.StoreKey
	QuerierRoute    = types.QuerierRoute
	ModuleName      = types.ModuleName
	QueryAccountMix = types.QueryAccountMix

	CodeSpaceAuthX           = types.CodeSpaceAuthX
	CodeGasPriceTooLow       = types.CodeGasPriceTooLow
	CodeRefereeChangeTooFast = types.CodeRefereeChangeTooFast

	DefaultParamspace       = types.DefaultParamspace
	DefaultMinGasPriceLimit = types.DefaultMinGasPriceLimit
)

var (
	ErrInvalidMinGasPriceLimit = types.ErrInvalidMinGasPriceLimit
	ErrGasPriceTooLow          = types.ErrGasPriceTooLow
	ErrRefereeChangeTooFast    = types.ErrRefereeChangeTooFast
	NewLockedCoin              = types.NewLockedCoin
	NewSupervisedLockedCoin    = types.NewSupervisedLockedCoin
	NewParams                  = types.NewParams
	NewAccountX                = types.NewAccountX
	DefaultParams              = types.DefaultParams
	ModuleCdc                  = types.ModuleCdc
	NewAccountXWithAddress     = types.NewAccountXWithAddress
	NewKeeper                  = keepers.NewKeeper
)

type (
	AccountX              = types.AccountX
	LockedCoin            = types.LockedCoin
	LockedCoins           = types.LockedCoins
	MsgSetReferee         = types.MsgSetReferee
	AccountXKeeper        = keepers.AccountXKeeper
	ExpectedAccountKeeper = keepers.ExpectedAccountKeeper
	ExpectedTokenKeeper   = keepers.ExpectedTokenKeeper
)
