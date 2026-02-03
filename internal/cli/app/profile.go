package app

import (
	"errors"
	"fmt"
	"slices"
	"stocks_calculator/internal/cli/client"
	"strings"
)

type ProfileOptions struct {
	Remove bool
	Info   bool
}

func Profile(profile string, options *ProfileOptions) (string, error) {

	if options.Info {
		return profileInfo(), nil
	}

	ind := slices.Index(userConfig.Profiles, profile)

	if options.Remove {
		return removeProfile(profile, ind)
	}

	return setDefaultOrCreate(profile, ind)
}

func profileInfo() string {
	if len(userConfig.Profiles) == 0 {
		return "No profiles have been created yet. Consider creating a profile to use stcalc."
	}

	var builder strings.Builder

	builder.WriteString("Available profiles:")

	for _, p := range userConfig.Profiles {
		if p == userConfig.DefaultProfile {
			fmt.Fprintf(&builder, "\n\t- %s (default)", p)
		} else {
			fmt.Fprintf(&builder, "\n\t- %s", p)
		}
	}

	if userConfig.DefaultProfile == "" {
		builder.WriteString("\n\nDefault profile is not set.")
	}

	return builder.String()
}

func removeProfile(profile string, ind int) (string, error) {
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
	if userConfig.DefaultProfile == profile {
		userConfig.DefaultProfile = ""
		message += " Default profile has been cleared. To use stcalc tool consider setting new default profile"
	}
	return message, updateConfig()
}

func setDefaultOrCreate(profile string, ind int) (string, error) {
	if ind != -1 {
		userConfig.DefaultProfile = profile
		return "Default profile has been set.", updateConfig()
	}

	err := client.CreateProfile(profile)
	if err != nil {
		return "", fmt.Errorf("Failed to create profile %q.", profile)
	}

	userConfig.Profiles = append(userConfig.Profiles, profile)

	message := fmt.Sprintf("Profile %q has been created.", profile)
	if userConfig.DefaultProfile == "" {
		userConfig.DefaultProfile = profile
		message += fmt.Sprintf(" Default profile has been set to %q.", profile)
	}

	return message, updateConfig()
}
