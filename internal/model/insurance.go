package model

type Insurance struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	CompanyName    string  `json:"company_name"`
	AmountRUB      float64 `json:"amount_rub"`
	AnnualInterest float64 `json:"annual_interest"`
	EndDate        string  `json:"end_date"`
}

type InsuranceUpdate struct {
	ID             int64    `json:"id"`
	Name           *string  `json:"name,omitempty"`
	CompanyName    *string  `json:"company_name,omitempty"`
	AnnualInterest *float64 `json:"annual_interest,omitempty"`
	EndDate        *string  `json:"end_date,omitempty"`
}

type InsuranceClose struct {
	ID               int64   `json:"id"`
	ClosingAmountRUB float64 `json:"closing_amount_rub"`
}
