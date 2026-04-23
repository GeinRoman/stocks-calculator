package model

import "time"

type (
	AuthModel struct {
		User string `json:"user"`
		Pass string `json:"pass"`
	}

	AuthResponse struct {
		Profiles []Profile `json:"profiles"`
		RefToken string    `json:"ref_token"`
		Token    `json:"token"`
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
		UserId int `json:"user_id"`
	}
)
