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

		GetGroups(ctx context.Context, userId int) ([]model.Group, error)
		InsertGroups(ctx context.Context, userId int, groups []model.Group) error
		UpdateGroupName(ctx context.Context, userId int, nameOld, nameNew string) error
		RemoveGroups(ctx context.Context, userId int, groups []model.Group) error
		UpdateWeights(ctx context.Context, userId int, groups []model.Group) error

		InsertStocks(ctx context.Context, userId int, stocks []model.Stock) error
		UpdateStocksAmount(ctx context.Context, userId int, stocks []model.Stock) error
		RemoveStocks(ctx context.Context, userId int, stocks []model.Stock) error
		GetStocks(ctx context.Context, userId int) ([]model.Stock, error)
	}
	TokenManager interface {
		GenerateJWT(userId int) (model.Token, error)
		GenerateRefreshToken() (string, error)
	}

	app struct {
		repo         Repository
		tokenManager TokenManager
	}
)

func New(repo Repository, tokenManager TokenManager) *app {
	return &app{repo: repo, tokenManager: tokenManager}
}
