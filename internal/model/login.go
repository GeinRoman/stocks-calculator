package model

import "time"

type LoginModel struct {
	User string
	Pass string
	New  bool
}

type LoginResponse struct {
	DefaultProfile string
	Profiles       []string
	RefToken       string
	Token
}

type Token struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
