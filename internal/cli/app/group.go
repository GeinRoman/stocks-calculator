package app

import (
	"fmt"
	"stocks_calculator/internal/cli/client"
)

func Add(groups []string) (string, error) {
	err := client.AddGroups(groups)
	if err != nil {
		return "", fmt.Errorf("Failed to add groups %v", groups)
	}

	if len(groups) == 1 {
		return "Group added successfully", nil
	}

	return "Groups added successfully", nil
}
