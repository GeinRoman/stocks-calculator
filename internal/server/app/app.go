package app

import (
	"context"
	"stocks_calculator/internal/model"
)

type (
	Repository interface {
		FindUserByUsername(ctx context.Context, user string) (model.User, error)
		CreateNewUser(ctx context.Context, user, hash string) (int, error)
		InsertRefreshToken(ctx context.Context, userId int, ref string) error
		UserIdByRefreshToken(ctx context.Context, refToken string) (int, error)

		RemoveProfile(ctx context.Context, profile string, userId int) error
		InsertProfile(ctx context.Context, profile string, userId int, def bool) error
		GetProfiles(ctx context.Context, userId int) ([]model.Profile, error)
		SetDefaultProfile(ctx context.Context, profile string, userId int) error

		GetGroups(ctx context.Context, userId int) ([]model.StockGroup, error)
		InsertGroups(ctx context.Context, userId int, groups []model.StockGroup) error
		UpdateGroupName(ctx context.Context, userId int, nameOld, nameNew string) error
		RemoveGroups(ctx context.Context, userId int, groups []model.StockGroup) error
		UpdateWeights(ctx context.Context, userId int, groups []model.StockGroup) error

		InsertStocks(ctx context.Context, userId int, stocks []model.Stock) error
		UpdateStocksAmount(ctx context.Context, userId int, stocks []model.Stock, amountDiff []int) error
		RemoveStocks(ctx context.Context, userId int, stocks []model.Stock) error
		GetStocks(ctx context.Context, userId int) ([]model.Stock, error)

		GetBondGroups(ctx context.Context, userId int) ([]model.BondGroup, error)
		InsertBondGroups(ctx context.Context, userId int, groups []model.BondGroup) error
		RemoveBondGroups(ctx context.Context, userId int, groups []model.BondGroup) error
		UpdateBondGroupName(ctx context.Context, userId int, nameOld, nameNew string) error
		UpdateBondWeights(ctx context.Context, userId int, groups []model.BondGroup) error

		InsertBonds(ctx context.Context, userId int, bonds []model.BondOperation) error
		UpdateBondsAmount(ctx context.Context, userId int, bonds []model.BondOperation, amountDiff []int) error
		RemoveBonds(ctx context.Context, userId int, bonds []model.BondOperation) error
		GetBonds(ctx context.Context, userId int) ([]model.Bond, error)

		GetCryptoGroups(context.Context, int) ([]model.CryptoGroup, error)
		InsertCryptoGroups(context.Context, int, []model.CryptoGroup) error
		RemoveCryptoGroups(context.Context, int, []model.CryptoGroup) error
		UpdateCryptoGroupName(context.Context, int, string, string) error
		UpdateCryptoWeights(context.Context, int, []model.CryptoGroup) error
		GetCryptoAssets(context.Context, int) ([]model.CryptoAsset, error)
		UpdateCryptoAsset(context.Context, int, model.CryptoAssetUpdate) error
		TransactCrypto(context.Context, int, []model.CryptoOperation, bool) error

		PersonalAssetProfileContext(context.Context, int) (context.Context, error)
		GetDeposits(context.Context, int) ([]model.Deposit, error)
		AddDeposit(context.Context, int, model.Deposit) error
		UpdateDeposit(context.Context, int, model.DepositUpdate) error
		RemoveDeposit(context.Context, int, int64) error
		CloseDeposit(context.Context, int, model.DepositClose) error
		GetInsurances(context.Context, int) ([]model.Insurance, error)
		AddInsurance(context.Context, int, model.Insurance) error
		UpdateInsurance(context.Context, int, model.InsuranceUpdate) error
		RemoveInsurance(context.Context, int, int64) error
		CloseInsurance(context.Context, int, model.InsuranceClose) error
		GetCreditCards(context.Context, int) ([]model.CreditCard, error)
		AddCreditCard(context.Context, int, model.CreditCard) error
		UpdateCreditCard(context.Context, int, model.CreditCardUpdate) error
		RemoveCreditCard(context.Context, int, int64) error
	}
	TokenManager interface {
		GenerateJWT(userId int) (model.Token, error)
		GenerateRefreshToken() (string, error)
	}
	MoexApi interface {
		FindStocks(ctx context.Context, names []string) []model.MoexFindResult
		GetInstrumentsInfo(ctx context.Context, codes []string) []model.MoexInstrumentInfoResult
		FindBonds(ctx context.Context, names []string) []model.MoexBondFindResult
		GetBondsInfo(ctx context.Context, codes []string) []model.MoexBondInfoResult
		GetBondCoupons(ctx context.Context, code string) (model.BondCouponsResponse, error)
		GetCurrencyRates(ctx context.Context) (map[string]float64, error)
	}

	app struct {
		repo         Repository
		tokenManager TokenManager
		moex         MoexApi
	}
)

func New(repo Repository, tokenManager TokenManager, moex MoexApi) *app {
	return &app{repo: repo, tokenManager: tokenManager, moex: moex}
}
