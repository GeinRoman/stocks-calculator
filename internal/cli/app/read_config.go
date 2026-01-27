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
)

type (
	userInfo struct {
		DefaultProfile string        `json:"default_profile"`
		Profiles       map[string]profileInfo `json:"profiles"`
	}

	profileInfo struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		ExpiresAt    time.Time `json:"expires_at"`
	}
)

var (
	userConfig     userInfo
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
		userConfig = userInfo{}
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
