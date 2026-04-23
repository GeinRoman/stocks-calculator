package model

type (
	Group struct {
		Name   string `json:"name"`
		Weight int    `json:"weight"`
	}

	RenameGroupBody struct {
		NameOld string `json:"name_old"`
		NameNew string `json:"name_new"`
	}
)
