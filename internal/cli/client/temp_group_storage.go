package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

const (
	dir       = "/tmp/stcalc"
	groupFile = "/tmp/stcalc/groups.json"
)

func readGroups() (groups []Group) {
	data, err := os.ReadFile(groupFile)
	if err != nil {
		return []Group{}
	}
	_ = json.Unmarshal(data, &groups)
	return groups
}

func writeGroups(groups []Group) {
	_, err := os.Stat(dir)
	if err != nil {
		os.Mkdir(dir, 0777)
	}

	data, _ := json.Marshal(groups)
	os.WriteFile(groupFile, data, 0666)
}

func groupExists(group *string) bool {
	groups := readGroups()
	ind, err := strconv.Atoi(*group)
	if err == nil {
		if ind > 0 && ind <= len(groups) {
			*group = groups[ind-1].Name
			return true
		} else {
			return false
		}
	}

	for i := range groups {
		if groups[i].Name == *group {
			return true
		}
	}
	return false
}
