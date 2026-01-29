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
	Remove         bool
	DefaultProfile bool
}

func Profile(profile string, options *ProfileOptions) (string, error) {

	if options.DefaultProfile {
		return defaultProfile()
	}

	_, exists := userConfig.Profiles[profile]

	if options.Remove {
		return removeProfile(profile, exists)
	}

	return setDefaultOrCreate(profile, exists)
}

func defaultProfile() (string, error) {
	if userConfig.DefaultProfile == "" {
		return "Default profile is not set", nil
	}
	return fmt.Sprintf("Default profile is %q.", userConfig.DefaultProfile), nil
}

func removeProfile(profile string, exists bool) (string, error) {
	if !exists {
		return "", fmt.Errorf("Profile %q does not exist.", profile)
	}

	delete(userConfig.Profiles, profile)

	message := fmt.Sprintf("Profile %q was removed.", profile)
	if userConfig.DefaultProfile == profile {
		userConfig.DefaultProfile = ""
		message += " Default profile has been cleared. To use stcalc tool consider setting new default profile"
	}
	return message, updateConfig()
}

func setDefaultOrCreate(profile string, exists bool) (string, error) {
	if exists {
		userConfig.DefaultProfile = profile
		return "Default profile has been set.", updateConfig()
	}

	// TODO: change defaults after adding token generation
	_, err := askPassword()
	if err != nil {
		return "", errors.New("Program failed to read password.")
	}
	// send password to service and retrieve tokens and time
	userConfig.Profiles[profile] = profileInfo{
		"",
		"",
		time.Now(),
	}

	message := fmt.Sprintf("Profile %q has been created.", profile)
	if len(userConfig.Profiles) == 1 {
		userConfig.DefaultProfile = profile
		message += fmt.Sprintf(" Default profile has been set to %q.", profile)
	}

	return message, updateConfig()
}

func askPassword() (string, error) {
	fmt.Print("Please, enter password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Print(clearLine)
	return string(password), err
}

func updateConfig() error {
	data, err := json.Marshal(userConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, data, 0600)
	return err
}
