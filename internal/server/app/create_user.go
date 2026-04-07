package app

import (
	"context"
	"stocks_calculator/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func (a *app) CreateUser(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Pass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id, err := a.repo.CreateNewUser(ctx, data.User, string(hash))
	if err != nil {
		return nil, err
	}

	token, err := a.tokenManager.GenerateJWT(id)
	if err != nil {
		return nil, err
	}
	refToken, err := a.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = a.repo.InsertRefreshToken(ctx, id, refToken)
	if err != nil {
		return nil, err
	}

	profiles, err := a.repo.Profiles(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Profiles: profiles,
		RefToken: refToken,
		Token:    token,
	}, nil
}
