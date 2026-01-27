package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

const clearLine = "\r\033[K"

type ProfileOptions struct {
	NewAccount bool
}

func Profile(profile string, options *ProfileOptions) (string, error) {
	password, err := askPassword()
	if err != nil {
		return "", err
	}

	err = saveToConfig(profile)
	if err != nil {
		return "", errors.New("Faild to save config file (" + err.Error() + ")")
	}

	return "logged in successfully with password: " + password + "", nil
}

func saveToConfig(profile string) error {
	userConfig = userInfo{
		DefaultProfile: profile,
		Profiles: map[string]profileInfo{
			profile: {"", "", time.Now()},
		},
	}

	data, err := json.Marshal(userConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, data, 0600)
	return err
}

func askPassword() (string, error) {
	fmt.Print("Please, enter password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Print(clearLine)
	return string(password), err
}
