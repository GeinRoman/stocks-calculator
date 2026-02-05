package client

import (
	"encoding/json"
	"os"
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
