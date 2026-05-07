package model

type (
	RebalanceResponse struct {
		Stocks        []Stock `json:"stocks"`
		TotalPrevCost float64 `json:"total_prev_cost"`
		CostDiff      float64 `json:"cost_diff"`
	}
)
