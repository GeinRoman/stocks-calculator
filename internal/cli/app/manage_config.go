package app

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
	configFileName = "user_profiles"
	defaultPort    = 8989
)

var (
	defaultConfig = userInfo{
		ConnectionStr: "",
		Port:          defaultPort,
		Username:      "",
		Profiles:      []model.Profile{},
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

func updateConfig() error {
	data, err := json.Marshal(userConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, data, 0600)
	return err
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
		userConfig = defaultConfig
		return nil
	case err != nil:
		return fmt.Errorf("Faild to open config file (%w)", err)
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("Faild to read config file (%w)", err)
	}

	err = json.Unmarshal(data, &userConfig)
	if err != nil {
		return fmt.Errorf("Faild to read config file (%w)", err)
	}
	return nil
}

func updateConfigAfterLogin(
	user string,
	response *model.AuthResponse,
) error {
	userConfig.Username = user
	userConfig.Profiles = response.Profiles
	userConfig.RefToken = response.RefToken
	userConfig.Token = response.Token
	return updateConfig()
}
