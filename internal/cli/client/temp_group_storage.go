package client

import (
	"encoding/json"
	"os"
)

const (
	dir  = "/tmp/stcalc"
	file = "/tmp/stcalc/groups.json"
)

func readGroups() (groups []Group) {
	data, err := os.ReadFile(file)
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
	os.WriteFile(file, data, 0666)
}
