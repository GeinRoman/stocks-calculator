package model

type Deposit struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	BankName       string  `json:"bank_name"`
	AmountRUB      float64 `json:"amount_rub"`
	AnnualInterest float64 `json:"annual_interest"`
	EndDate        string  `json:"end_date"`
}

type DepositUpdate struct {
	ID             int64    `json:"id"`
	Name           *string  `json:"name,omitempty"`
	BankName       *string  `json:"bank_name,omitempty"`
	AnnualInterest *float64 `json:"annual_interest,omitempty"`
	EndDate        *string  `json:"end_date,omitempty"`
}

type DepositClose struct {
	ID               int64   `json:"id"`
	ClosingAmountRUB float64 `json:"closing_amount_rub"`
}
