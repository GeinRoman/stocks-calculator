package app

import (
	"errors"
	"fmt"
	"slices"
	"stocks_calculator/internal/cli/client"
	"stocks_calculator/internal/model"
	"strings"
)

type ProfileOptions struct {
	Remove bool
	Info   bool
}

func Profile(profileName string, options *ProfileOptions) (string, error) {

	if options.Info {
		return profileInfo(), nil
	}

	ind := -1
	indDefault := -1
	for i, p := range userConfig.Profiles {
		if p.Name == profileName {
			ind = i
		}
		if p.Default {
			indDefault = i
		}
	}

	if options.Remove {
		return removeProfile(profileName, ind, indDefault)
	}

	return setDefaultOrCreate(profileName, ind, indDefault)
}

func profileInfo() string {
	if len(userConfig.Profiles) == 0 {
		return "No profiles have been created yet. Consider creating a profile to use stcalc."
	}

	var builder strings.Builder

	builder.WriteString("Available profiles:")

	hasDefault := false
	for _, p := range userConfig.Profiles {
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

	return builder.String()
}

func removeProfile(profile string, ind, indDefault int) (string, error) {
	if ind == -1 {
		return "", fmt.Errorf("Profile %q does not exist.", profile)
	}

	confirmed, err := confirmation(fmt.Sprintf("Are you sure you want to remove profile %q?", profile))
	if err != nil {
		return "", errors.New("Fail to read user input")
	}
	if !confirmed {
		return "Profile was not removed", nil
	}

	err = client.RemoveProfile(profile)
	if err != nil {
		return "", fmt.Errorf("Failed to remove profile %q.", profile)
	}

	userConfig.Profiles = slices.Delete(userConfig.Profiles, ind, ind+1)

	message := fmt.Sprintf("Profile %q was removed.", profile)
	if ind == indDefault {
		message += " Default profile has been cleared. To use stcalc tool consider setting new default profile"
	}

	return message, updateConfig()
}

func setDefaultOrCreate(profile string, ind, indDefault int) (string, error) {
	if ind != -1 {
		if indDefault != -1 {
			userConfig.Profiles[indDefault].Default = false
		}
		userConfig.Profiles[ind].Default = true
		return "Default profile has been set.", updateConfig()
	}

	err := client.CreateProfile(profile)
	if err != nil {
		return "", fmt.Errorf("Failed to create profile %q.", profile)
	}

	userConfig.Profiles = append(userConfig.Profiles, model.Profile{Name: profile, Default: false})

	message := fmt.Sprintf("Profile %q has been created.", profile)
	if indDefault != 1 {
		userConfig.Profiles[len(userConfig.Profiles)-1].Default = true
		message += fmt.Sprintf(" Default profile has been set to %q.", profile)
	}

	return message, updateConfig()
}
