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
