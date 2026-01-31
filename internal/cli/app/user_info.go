package app

import (
	"time"
)

type (
	userInfo struct {
		ConnectionStr  string                 `json:"connection_str"`
		Port           int                    `json:"port"`
		DefaultProfile string                 `json:"default_profile"`
		Profiles       map[string]profileInfo `json:"profiles"`
	}

	profileInfo struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		ExpiresAt    time.Time `json:"expires_at"`
	}
)

var userConfig userInfo
