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
	AccessToken    string
	RefreshToken   string
	ExpiresAt      time.Time
}
