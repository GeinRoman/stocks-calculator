package model

import "time"

type TransactionCursor struct {
	Before   *time.Time
	BeforeID *int64
}

type DepositTransaction struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	BankName  string    `json:"bank_name"`
	AmountRUB float64   `json:"amount_rub"`
	Buying    bool      `json:"buying"`
	CreatedAt time.Time `json:"created_at"`
}

type InsuranceTransaction struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	CompanyName string    `json:"company_name"`
	AmountRUB   float64   `json:"amount_rub"`
	Buying      bool      `json:"buying"`
	CreatedAt   time.Time `json:"created_at"`
}

type CryptoTransaction struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	AssetName string    `json:"asset_name"`
	Platform  string    `json:"platform"`
	Amount    float64   `json:"amount"`
	Buying    bool      `json:"buying"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type StockTransaction struct {
	ID             int64     `json:"id"`
	InstrumentCode string    `json:"instrument_code"`
	LotAmount      int       `json:"lot_amount"`
	Buying         bool      `json:"buying"`
	LotSize        int       `json:"lot_size"`
	Price          float64   `json:"price"`
	Currency       string    `json:"currency"`
	Time           time.Time `json:"time"`
}

type BondTransaction StockTransaction
