package app

import (
	"fmt"
	"stocks_calculator/internal/cli/client"
	"stocks_calculator/internal/model"
)

type LoginOptions struct {
	User string
	New  bool
}

func Login(options LoginOptions) (string, error) {
	if options.User == "" {
		if userConfig.Username == "" {
			return "You are not currently logged in.", nil
		}
		return fmt.Sprintf("You are logged in as %q.", userConfig.Username), nil
	}

	pass, err := getPassword(options.New)
	if err != nil {
		return "", err
	}

	response, err := client.Login(model.LoginModel{
		User: options.User,
		Pass: pass,
		New:  options.New,
	})
	if err != nil {
		return "", fmt.Errorf("Failed to login on remote server (%w)", err)
	}

	if err = updateConfigAfterLogin(options.User, &response); err != nil {
		return "", fmt.Errorf("Failed to update config file (%w)", err)
	}

	var output string
	if options.New {
		output = fmt.Sprintf("Successfully logged in as %q.", options.User)
	} else {
		output = fmt.Sprintf("Successfully created user %q and logged in.", options.User)
	}
	return output, nil
}

func getPassword(newPass bool) (pass string, err error) {
	if newPass {
		for {
			pass, err = askPassword("Create new password")
			if err != nil {
				return "", fmt.Errorf("Failed to read password (%w)", err)
			}
			var conf string
			conf, err = askPassword("Confirm password")
			if err != nil {
				return "", fmt.Errorf("Failed to read password (%w)", err)
			}
			if pass == conf {
				return
			}

			fmt.Println("Failed to confirm password, please try again.")
		}
	}

	pass, err = askPassword("Please, enter password")
	if err != nil {
		return "", fmt.Errorf("Failed to read password (%w)", err)
	}

	return
}
