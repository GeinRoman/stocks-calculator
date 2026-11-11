package model

import "errors"

var (
	ErrInvalidCreditCard  = errors.New("invalid credit card")
	ErrCreditCardNotFound = errors.New("credit card not found")
)

type CreditCard struct {
	ID                        int64   `json:"id"`
	Name                      string  `json:"name"`
	BankName                  string  `json:"bank_name"`
	MandatoryPaymentDate      string  `json:"mandatory_payment_date"`
	MandatoryPaymentAmountRUB float64 `json:"mandatory_payment_amount_rub"`
	OverallLimitRUB           float64 `json:"overall_limit_rub"`
}

type CreditCardUpdate struct {
	ID                        int64    `json:"id"`
	Name                      *string  `json:"name,omitempty"`
	BankName                  *string  `json:"bank_name,omitempty"`
	MandatoryPaymentDate      *string  `json:"mandatory_payment_date,omitempty"`
	MandatoryPaymentAmountRUB *float64 `json:"mandatory_payment_amount_rub,omitempty"`
	OverallLimitRUB           *float64 `json:"overall_limit_rub,omitempty"`
}
