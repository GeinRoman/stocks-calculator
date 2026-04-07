package model

import "time"

type AuthModel struct {
	User string
	Pass string
}

type AuthResponse struct {
	Profiles []Profile
	RefToken string
	Token
}

type Token struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type Profile struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type User struct {
	Id       int
	Name     string
	Password string
}
