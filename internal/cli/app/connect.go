package app

import (
	"fmt"
	"stocks_calculator/internal/cli/config"
	"strconv"
)

type ConnectOptions struct {
	Port bool
	Info bool
}

func Connect(arg string, options *ConnectOptions) (string, error) {
	if options.Info {
		return showInfo(), nil
	}

	if options.Port {
		return updatePort(arg)
	}

	return updateConnectionStr(arg)
}

func showInfo() string {
	if config.UserConfig.ConnectionStr == "" {
		return "Connection string is empty. Please, set connection string before using stcalc"
	}

	return fmt.Sprintf("Connection string is: %s. Port is: %d", config.UserConfig.ConnectionStr, config.UserConfig.Port)
}

func updatePort(arg string) (string, error) {
	port, err := strconv.Atoi(arg)
	if err != nil {
		return "", fmt.Errorf("Invalid port: %s must be an integer", arg)
	}

	if port < 1024 || port > 65535 {
		return "", fmt.Errorf("Port must be between 1024-65535 (user port range)")
	}

	config.UserConfig.Port = port
	return fmt.Sprintf("Port successfully updated to %s", arg), config.UpdateConfig()
}

func updateConnectionStr(arg string) (string, error) {
	config.UserConfig.ConnectionStr = arg
	return fmt.Sprintf("Connection string successfully updated to %s", arg), config.UpdateConfig()
}
