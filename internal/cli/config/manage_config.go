package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"stocks_calculator/internal/model"
	"time"
)

const (
	appConfigDir   = "stcalc"
	configFileName = "config"
	defaultPort    = 8989
)

var (
	defaultConfig = userInfo{
		ConnectionStr: "",
		Port:          defaultPort,
		Username:      "",
		RefToken:      "",
		Token: model.Token{
			AccessToken: "",
			ExpiresAt:   time.Now(),
		},
	}
	configFilePath string
)

func ReadConfig() error {
	dir, err := detectConfigDir()
	if err != nil {
		return fmt.Errorf("Faild to determine config directory")
	}

	dir = filepath.Join(dir, appConfigDir)
	configFilePath = filepath.Join(dir, configFileName)

	err = createAppDir(dir)
	if err != nil {
		return fmt.Errorf("Faild to create config dir (%w)", err)
	}

	return readFileToUserConfig()
}

func UpdateConfig() error {
	data, err := json.Marshal(UserConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, data, 0600)
	return err
}

func UpdateAccessToken(token model.Token) error {
	UserConfig.Token = token
	return UpdateConfig()
}

func detectConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err == nil {
		return dir, nil
	}
	return os.UserHomeDir()
}

func createAppDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0700)
	}
	return err
}

func readFileToUserConfig() error {
	_, err := os.Stat(configFilePath)
	switch {
	case os.IsNotExist(err):
		UserConfig = defaultConfig
		return nil
	case err != nil:
		return fmt.Errorf("Faild to open config file (%w)", err)
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("Faild to read config file (%w)", err)
	}

	err = json.Unmarshal(data, &UserConfig)
	if err != nil {
		return fmt.Errorf("Faild to read config file (%w)", err)
	}
	return nil
}

func UpdateConfigAfterLogin(
	user string,
	response model.AuthResponse,
) error {
	UserConfig.Username = user
	UserConfig.RefToken = response.RefToken
	UserConfig.Token = response.Token
	return UpdateConfig()
}
