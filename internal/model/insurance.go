package model

type Insurance struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	CompanyName    string  `json:"company_name"`
	AmountRUB      float64 `json:"amount_rub"`
	AnnualInterest float64 `json:"annual_interest"`
	StartDate      *string `json:"start_date"`
	EndDate        string  `json:"end_date"`
}

type InsuranceUpdate struct {
	ID             int64    `json:"id"`
	Name           *string  `json:"name,omitempty"`
	CompanyName    *string  `json:"company_name,omitempty"`
	AnnualInterest *float64 `json:"annual_interest,omitempty"`
	StartDate      *string  `json:"start_date,omitempty"`
	EndDate        *string  `json:"end_date,omitempty"`
}

type InsuranceClose struct {
	ID               int64   `json:"id"`
	ClosingAmountRUB float64 `json:"closing_amount_rub"`
}
