package model

type (
	Profile struct {
		Name    string `json:"name"`
		Default bool   `json:"default"`
	}

	RenameProfileBody struct {
		Name    string `json:"name"`
		NewName string `json:"new_name"`
	}
)
