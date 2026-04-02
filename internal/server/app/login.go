package app

import (
	"stocks_calculator/internal/model"
	"time"
)

func Login(data model.LoginModel) (model.LoginResponse, error) {
	if data.User == "" || data.Pass == "" {
		return model.LoginResponse{}, ErrWrongCredentials
	}

	return model.LoginResponse{
		DefaultProfile: "aboba",
		Profiles:       []string{"aboba"},
		RefToken:       "312",
		Token: model.Token{
			AccessToken: "123",
			ExpiresAt:   time.Now(),
		},
	}, nil
}
