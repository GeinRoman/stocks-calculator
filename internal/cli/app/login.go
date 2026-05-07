package app

import (
	"context"
	"fmt"
	"stocks_calculator/internal/cli/client"
	"stocks_calculator/internal/cli/config"
	"stocks_calculator/internal/model"
	"time"
)

type LoginOptions struct {
	User string
	New  bool
}

func Login(options LoginOptions) (string, error) {
	if options.User == "" {
		if config.UserConfig.Username == "" {
			return "You are not currently logged in.", nil
		}
		return fmt.Sprintf("You are logged in as %q.", config.UserConfig.Username), nil
	}

	pass, err := getPassword(options.New)
	if err != nil {
		return "", err
	}

	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var response model.AuthResponse

	if options.New {
		response, err = httpClient.CreateUser(ctx, model.AuthModel{
			User: options.User,
			Pass: pass,
		})
	} else {
		response, err = httpClient.Login(ctx, model.AuthModel{
			User: options.User,
			Pass: pass,
		})
	}

	if err != nil {
		return "", fmt.Errorf("Failed to login on remote server (%w)", err)
	}

	if err = config.UpdateConfigAfterLogin(options.User, response); err != nil {
		return "", fmt.Errorf("Failed to update config file (%w)", err)
	}

	var output string
	if options.New {
		output = fmt.Sprintf("Successfully created user %q and logged in.", options.User)
	} else {
		output = fmt.Sprintf("Successfully logged in as %q.", options.User)
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
