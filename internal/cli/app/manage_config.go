package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	appConfigDir   = ".stcalc"
	configFileName = "user_profiles"
	defaultPort    = 8989
)

var (
	defaultConfig = userInfo{
		ConnectionStr:  "",
		Port:           defaultPort,
		DefaultProfile: "",
		Profiles:       []string{},
		AccessToken:    "",
		RefreshToken:   "",
		ExpiresAt:      time.Now(),
	}
	configFilePath string
)

func ReadConfig() error {
	dir, err := detectConfigDir()
	if err != nil {
		return errors.New("Faild to determine config directory")
	}

	dir = filepath.Join(dir, appConfigDir)
	configFilePath = filepath.Join(dir, configFileName)

	err = createAppDir(dir)
	if err != nil {
		return errors.New("Faild to create config dir (" + err.Error() + ")")
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
		return errors.New("Faild to open config file (" + err.Error() + ")")
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return errors.New("Faild to read config file (" + err.Error() + ")")
	}

	err = json.Unmarshal(data, &userConfig)
	if err != nil {
		return errors.New("Faild to read config file (" + err.Error() + ")")
	}
	return nil
}
