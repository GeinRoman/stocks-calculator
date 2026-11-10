package model

type BondCouponsResponse struct {
	Code    string   `json:"code"`
	Coupons []Coupon `json:"coupons"`
}

type Coupon struct {
	PaymentDate string   `json:"payment_date"` // YYYY-MM-DD.
	Amount      *float64 `json:"amount"`       // Gross RUB per bond; nil if unknown.
}
