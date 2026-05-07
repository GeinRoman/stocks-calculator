package app

import (
	"context"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"

	"golang.org/x/crypto/bcrypt"
)

func (a *app) Login(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error) {
	user, err := a.repo.FindUserByUsername(ctx, data.User)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Pass))
	if err != nil {
		return nil, servererrors.ErrWrongCredentials
	}

	token, err := a.tokenManager.GenerateJWT(user.Id)
	if err != nil {
		return nil, err
	}
	refToken, err := a.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = a.repo.InsertRefreshToken(ctx, user.Id, refToken)
	if err != nil {
		return nil, err
	}

	profiles, err := a.repo.GetProfiles(ctx, user.Id)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Profiles: profiles,
		RefToken: refToken,
		Token:    token,
	}, nil
}

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

	profiles, err := a.repo.GetProfiles(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Profiles: profiles,
		RefToken: refToken,
		Token:    token,
	}, nil
}

func (a *app) RefreshToken(ctx context.Context, refToken string) (*model.Token, error) {
	id, err := a.repo.UserIdByRefreshToken(ctx, refToken)
	if err != nil {
		return nil, err
	}
	token, err := a.tokenManager.GenerateJWT(id)
	return &token, nil
}
