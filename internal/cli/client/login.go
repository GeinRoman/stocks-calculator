package client

import (
	"stocks_calculator/internal/model"
	"time"
)

func Login(data model.AuthModel) (model.AuthResponse, error) {
	// sample return
	// placeholder for http request
	return model.AuthResponse{
		Profiles: []model.Profile{},
		RefToken: "312",
		Token: model.Token{
			AccessToken: "123",
			ExpiresAt:   time.Now(),
		},
	}, nil
}
