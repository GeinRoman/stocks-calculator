package model

type (
	StockGroup struct {
		Name   string `json:"name"`
		Weight int    `json:"weight"`
	}

	RenameStockGroupBody struct {
		NameOld string `json:"name_old"`
		NameNew string `json:"name_new"`
	}
)
