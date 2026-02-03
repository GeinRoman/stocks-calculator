package app

import (
	"errors"
	"fmt"
	"time"
)

type (
	userInfo struct {
		ConnectionStr  string    `json:"connection_str"`
		Port           int       `json:"port"`
		DefaultProfile string    `json:"default_profile"`
		Profiles       []string  `json:"profiles"`
		AccessToken    string    `json:"access_token"`
		RefreshToken   string    `json:"refresh_token"`
		ExpiresAt      time.Time `json:"expires_at"`
	}
)

var userConfig userInfo

func Url() (string, error) {
	if userConfig.ConnectionStr == "" {
		return "", errors.New("Connection string is empty. Please add connection string with \"stcalc connect\"")
	}

	return fmt.Sprintf("%s:%d", userConfig.ConnectionStr, userConfig.Port), nil
}
