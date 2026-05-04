package app

import (
	"fmt"
	"stocks_calculator/internal/cli/config"
)

func ValidateConnectionString() error {
	if config.UserConfig.ConnectionStr == "" {
		return fmt.Errorf("Connection string is not set. Please configure the connection using: \"stcalc connect <connection-str>\"")
	}
	return nil
}

func ValidateLogin() error {
	if config.UserConfig.RefToken == "" || config.UserConfig.Username == "" {
		return fmt.Errorf("You are not logged in. Please log in using: \"stcalc login -u <username>\"")
	}
	return nil
}
