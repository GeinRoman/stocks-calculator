package client

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	Group struct {
		Name   string `json:"name"`
		Weight int    `json:"weight"`
	}
)

func AddGroups(groups []string) error {
	//temp groups saving functionality
	//placeholder for http request

	prevGroups := readGroups()

	groupsToAdd := []string{}
outer:
	for _, g := range groups {
		for _, pg := range prevGroups {
			if strings.EqualFold(g, pg.Name) {
				continue outer
			}
		}
		groupsToAdd = append(groupsToAdd, g)
	}

	for _, name := range groupsToAdd {
		prevGroups = append(prevGroups, Group{name, 0})
	}

	writeGroups(prevGroups)

	return nil
}

type (
	RemoveBody struct {
		Names   []string `json:"names"`
		Indexes []int    `json:"indexes"`
	}
)

func RemoveGroups(groups *RemoveBody) error {
	//temp groups removing functionality
	//placeholder for http request
	prevGroups := readGroups()

	for _, ind := range groups.Indexes {
		if len(prevGroups) < ind {
			continue
		}

		prevGroups[ind-1].Name = ""
	}
	for _, name := range groups.Names {
		for i := range prevGroups {
			if prevGroups[i].Name == name {
				prevGroups[i].Name = ""
			}
		}
	}
	for i := 0; ; {
		if i == len(prevGroups) {
			break
		}

		if prevGroups[i].Name == "" {
			if i == len(prevGroups)-1 {
				prevGroups = prevGroups[:i]
			} else {
				prevGroups = append(prevGroups[:i], prevGroups[i+1:]...)
			}
			continue
		}

		i++
	}

	writeGroups(prevGroups)

	return nil
}

func RenameGroup(oldN string, newN string) error {
	//temp renaming functionality
	//placeholder for http request
	prevGroups := readGroups()

	renamed := false
	ind, err := strconv.Atoi(oldN)
	if err == nil {
		if ind > 0 && ind <= len(prevGroups) {
			prevGroups[ind-1].Name = newN
			renamed = true
		}
	} else {
		for i := range prevGroups {
			if prevGroups[i].Name == oldN {
				prevGroups[i].Name = newN
				renamed = true
				break
			}
		}
	}

	if !renamed {
		return fmt.Errorf("Group with name or index %q was not found", oldN)
	}

	writeGroups(prevGroups)
	return nil
}

func GetGroups() ([]Group, error) {
	//temp get groups functionality
	//placeholder for http request
	return readGroups(), nil
}

func UpdateWeights(groups []Group) error {
	//temp get groups functionality
	//placeholder for http request
	writeGroups(groups)
	return nil
}
