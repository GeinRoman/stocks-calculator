package config

import (
	"fmt"
	"stocks_calculator/internal/model"
)

type (
	userInfo struct {
		ConnectionStr string `json:"connection_str"`
		Port          int    `json:"port"`
		Username      string `json:"username"`
		RefToken      string `json:"ref_token"`
		model.Token   `json:"token"`
	}
)

var UserConfig userInfo

func Url() string {
	return fmt.Sprintf("%s:%d", UserConfig.ConnectionStr, UserConfig.Port)
}
