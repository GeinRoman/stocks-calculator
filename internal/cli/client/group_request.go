package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func AddGroups(groups []string) error {
	//temp groups saving functionality
	//placeholder for http request
	const (
		dir  = "/tmp/stcalc"
		file = "/tmp/stcalc/groups.json"
	)
	_, err := os.Stat(dir)
	if err != nil {
		os.Mkdir(dir, 0777)
	}

	_, err = os.Stat(file)
	if err != nil {
		data, _ := json.Marshal(groups)
		os.WriteFile(file, data, 0666)
	}

	data, _ := os.ReadFile(file)
	var prevGroups []string
	_ = json.Unmarshal(data, &prevGroups)

	groupsToAdd := []string{}
outer:
	for _, g := range groups {
		for _, pg := range prevGroups {
			if strings.EqualFold(g, pg) {
				continue outer
			}
		}
		groupsToAdd = append(groupsToAdd, g)
	}

	prevGroups = append(prevGroups, groupsToAdd...)
	data, _ = json.Marshal(prevGroups)
	os.WriteFile(file, data, 0666)

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
	const (
		file = "/tmp/stcalc/groups.json"
	)

	_, err := os.Stat(file)
	if err != nil {
		return nil
	}

	data, _ := os.ReadFile(file)
	var prevGroups []string
	_ = json.Unmarshal(data, &prevGroups)

	for _, ind := range groups.Indexes {
		if len(prevGroups) < ind {
			continue
		}

		prevGroups[ind-1] = ""
	}
	for _, name := range groups.Names {
		for i := range prevGroups {
			if prevGroups[i] == name {
				prevGroups[i] = ""
			}
		}
	}
	for i := 0; ; {
		if i == len(prevGroups) {
			break
		}

		if prevGroups[i] == "" {
			if i == len(prevGroups)-1 {
				prevGroups = prevGroups[:i]
			} else {
				prevGroups = append(prevGroups[:i], prevGroups[i+1:]...)
			}
			continue
		}

		i++
	}

	data, _ = json.Marshal(prevGroups)
	os.WriteFile(file, data, 0666)

	return nil
}

func RenameGroup(oldN string, newN string) error {
	//temp renaming functionality
	//placeholder for http request
	const (
		file = "/tmp/stcalc/groups.json"
	)

	_, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("Group with name or index %q was not found", oldN)
	}

	data, _ := os.ReadFile(file)
	var prevGroups []string
	_ = json.Unmarshal(data, &prevGroups)
	renamed := false

	ind, err := strconv.Atoi(oldN)
	if err == nil {
		if ind > 0 && ind <= len(prevGroups) {
			prevGroups[ind-1] = newN
			renamed = true
		}
	} else {
		for i := range prevGroups {
			if prevGroups[i] == oldN {
				prevGroups[i] = newN
				renamed = true
				break
			}
		}
	}

	if !renamed {
		return fmt.Errorf("Group with name or index %q was not found", oldN)
	}

	data, _ = json.Marshal(prevGroups)
	os.WriteFile(file, data, 0666)

	return nil
}
