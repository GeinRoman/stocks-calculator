package app

import (
	"context"
	"stocks_calculator/internal/model"
)

type (
	Repository interface {
		FindUserByUsername(ctx context.Context, user string) (model.User, error)
		CreateNewUser(ctx context.Context, user, hash string) (int, error)
		Profiles(ctx context.Context, userId int) ([]model.Profile, error)
		InsertRefreshToken(ctx context.Context, userId int, ref string) error
		UserIdByRefreshToken(ctx context.Context, refToken string) (int, error)
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
