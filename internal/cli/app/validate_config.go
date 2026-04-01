package app

import "fmt"

func ValidateProfile() error {
	switch {
	case len(userConfig.Profiles) == 0:
		return fmt.Errorf("No profiles found. Please create a profile first using: \"stcalc profile <profile-name>\"")
	case userConfig.DefaultProfile == "":
		return fmt.Errorf("No default profile set. Please set a default profile using: \"stcalc profile set-default <profile-name>\"")
	}

	return nil
}

func ValidateConnectionString() error {
	if userConfig.ConnectionStr == "" {
		return fmt.Errorf("Connection string is not set. Please configure the connection using: \"stcalc connect <connection-str>\"")
	}
	return nil
}

func ValidateLogin() error {
	if userConfig.AccessToken == "" {
		return fmt.Errorf("You are not logged in. Please log in using: \"stcalc login -u <username>\"")
	}
	return nil
}
