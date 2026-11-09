package model

type (
	StockRebalanceResponse struct {
		Stocks        []Stock `json:"stocks"`
		TotalPrevCost float64 `json:"total_prev_cost"`
		CostDiff      float64 `json:"cost_diff"`
	}

	BondRebalanceResponse struct {
		Bonds         []Bond  `json:"bonds"`
		TotalPrevCost float64 `json:"total_prev_cost"`
		CostDiff      float64 `json:"cost_diff"`
	}
)
