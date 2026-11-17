package model

type (
	Stock struct {
		Name      string  `json:"name"`
		Code      string  `json:"code"`
		GroupName string  `json:"group_name"`
		Price     float64 `json:"price"`
		LotAmount int     `json:"lot_amount"`
		LotSize   int     `json:"lot_size"`
	}

	FoundStock struct {
		SearchResults []Stock `json:"search_results"`
		ErrMsg        string  `json:"err_msg"`
	}
)
