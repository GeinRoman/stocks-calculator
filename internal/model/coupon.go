package model

type BondCouponsResponse struct {
	Code    string   `json:"code"`
	Coupons []Coupon `json:"coupons"`
}

type Coupon struct {
	PaymentDate string `json:"payment_date"` // YYYY-MM-DD.
	Amount      *Money `json:"amount"`       // Gross per bond; nil if unknown.
}
