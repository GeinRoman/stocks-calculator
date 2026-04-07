package app

import (
	"context"
	"stocks_calculator/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func (a *app) Login(ctx context.Context, data model.AuthModel) (*model.AuthResponse, error) {
	user, err := a.repo.FindUserByUsername(ctx, data.User)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Pass))
	if err != nil {
		return nil, ErrWrongCredentials
	}
	// hash, err := bcrypt.GenerateFromPassword([]byte(data.Pass), bcrypt.DefaultCost)
	// if err != nil {
	// 	return nil, err
	// }

	// var id int
	// if data.New {
	// 	id, err = a.repo.CreateNewUser(ctx, data.User, string(hash))
	// } else {
	// 	id, err = a.repo.FindUser(ctx, data.User, string(hash))
	// }

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

	profiles, err := a.repo.Profiles(ctx, user.Id)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Profiles: profiles,
		RefToken: refToken,
		Token:    token,
	}, nil
}
