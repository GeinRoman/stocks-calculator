package app

import "errors"

func ValidateUserConfig() error {
	switch {
	case len(userConfig.Profiles) == 0:
		return errors.New("No profiles found. Please create a profile first using: stcalc profile <profile-name>")
	case userConfig.DefaultProfile == "":
		return errors.New("No default profile set. Please set a default profile using: stcalc profile set-default <profile-name>")
	case userConfig.ConnectionStr == "":
		return errors.New("Connection string is not set. Please configure the connection using: stcalc connect <connection-str>")
	}

	return nil
}

