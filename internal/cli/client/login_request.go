package client

import (
	"stocks_calculator/internal/model"
	"time"
)

func Login(data model.LoginModel) (model.LoginResponse, error) {
	// sample return
	// placeholder for http request
	return model.LoginResponse{
		DefaultProfile: "",
		Profiles:       []string{},
		AccessToken:    "123",
		RefreshToken:   "321",
		ExpiresAt:      time.Now(),
	}, nil
}
