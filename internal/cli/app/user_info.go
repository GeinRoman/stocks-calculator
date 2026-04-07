package app

import (
	"fmt"
	"stocks_calculator/internal/model"
)

type (
	userInfo struct {
		ConnectionStr string          `json:"connection_str"`
		Port          int             `json:"port"`
		Username      string          `json:"username"`
		Profiles      []model.Profile `json:"profiles"`
		RefToken      string          `json:"ref_token"`
		model.Token   `json:"token"`
	}
)

var userConfig userInfo

func Url() (string, error) {
	if userConfig.ConnectionStr == "" {
		return "", fmt.Errorf("Connection string is empty. Please add connection string with \"stcalc connect\"")
	}

	return fmt.Sprintf("%s:%d", userConfig.ConnectionStr, userConfig.Port), nil
}
