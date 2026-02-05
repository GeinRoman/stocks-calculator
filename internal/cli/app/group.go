package app

import (
	"fmt"
	"stocks_calculator/internal/cli/client"
	"strconv"
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

func Remove(groups []string) (string, error) {
	toRemove := client.RemoveBody{
		Names:   []string{},
		Indexes: []int{},
	}

	for _, v := range groups {
		if ind, err := strconv.Atoi(v); err == nil {
			if ind > 0 {
				toRemove.Indexes = append(toRemove.Indexes, ind)
			}
		} else {
			toRemove.Names = append(toRemove.Names, v)
		}
	}

	err := client.RemoveGroups(&toRemove)

	if err != nil {
		return "", fmt.Errorf("Failed to remove groups %v", groups)
	}

	if len(groups) == 1 {
		return "Group removed successfully", nil
	}

	return "Groups removed successfully", nil
}
