package app

import (
	"context"
	"errors"
	"fmt"
	"stocks_calculator/internal/cli/client"
	"stocks_calculator/internal/cli/config"
	"stocks_calculator/internal/model"
	"strings"
	"time"
)

type ProfileOptions struct {
	ProfileName string
	Remove      bool
	Info        bool
}

func Profile(options ProfileOptions) (string, error) {
	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if options.Info {
		return profileInfo(ctx, httpClient)
	}

	if options.Remove {
		return removeProfile(ctx, httpClient, options.ProfileName)
	}

	return setDefaultOrCreate(ctx, httpClient, options.ProfileName)
}

func profileInfo(ctx context.Context, httpClient *client.HttpClient) (string, error) {
	profiles, err := httpClient.GetProfiles(ctx)
	if err != nil {
		return "", err
	}

	if len(profiles) == 0 {
		return "No profiles have been created yet. Consider creating a profile to use stcalc.", nil
	}

	var builder strings.Builder

	builder.WriteString("Available profiles:")

	hasDefault := false
	for _, p := range profiles {
		if p.Default {
			fmt.Fprintf(&builder, "\n\t- %s (default)", p.Name)
			hasDefault = true
		} else {
			fmt.Fprintf(&builder, "\n\t- %s", p.Name)
		}
	}

	if !hasDefault {
		builder.WriteString("\n\nDefault profile is not set.")
	}

	return builder.String(), nil
}

func removeProfile(ctx context.Context, httpClient *client.HttpClient, profile string) (string, error) {
	confirmed, err := confirmation(fmt.Sprintf("Are you sure you want to remove profile %q?", profile))
	if err != nil {
		return "", errors.New("Fail to read user input")
	}
	if !confirmed {
		return "Profile was not removed", nil
	}

	err = httpClient.RemoveProfile(ctx, model.Profile{Name: profile})
	if err != nil {
		return "", fmt.Errorf("Failed to remove profile %q. %w", profile, err)
	}

	return fmt.Sprintf("Profile %q was removed.", profile), nil
}

func setDefaultOrCreate(ctx context.Context, httpClient *client.HttpClient, profile string) (string, error) {
	profiles, err := httpClient.GetProfiles(ctx)
	if err != nil {
		return "", err
	}

	for _, p := range profiles {
		if p.Name == profile {
			if p.Default {
				return fmt.Sprintf("Default profile has been set to %q.", profile), nil
			}
			err = httpClient.SetDefaultProfile(ctx, model.Profile{Name: profile, Default: true})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Default profile has been set to %q.", profile), nil
		}
	}

	err = httpClient.CreateProfile(ctx, model.Profile{Name: profile, Default: true})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Profile %q has been created and chosen as default.", profile), nil
}
