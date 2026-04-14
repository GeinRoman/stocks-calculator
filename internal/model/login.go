package model

import "time"

type (
	AuthModel struct {
		User string
		Pass string
	}

	AuthResponse struct {
		Profiles []Profile
		RefToken string
		Token
	}

	User struct {
		Id       int
		Name     string
		Password string
	}

	Token struct {
		AccessToken string    `json:"access_token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}

	Claims struct {
		UserId int
	}
)
