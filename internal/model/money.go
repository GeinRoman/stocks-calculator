package model

type Money struct {
	Currency string  `json:"currency"`
	Price    float64 `json:"price"`
}

type CurrencyRate struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rate float64 `json:"rate"`
}
