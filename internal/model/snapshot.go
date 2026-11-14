package model

import "time"

// Snapshot IDs are file-local references, not database IDs to restore verbatim.
type Snapshot struct {
	Version  int               `json:"version"`
	Profiles []SnapshotProfile `json:"profiles"`
}
type SnapshotProfile struct {
	Name                  string                      `json:"name"`
	Default               bool                        `json:"default"`
	StockGroups           []SnapshotStockGroup        `json:"stock_groups"`
	BondGroups            []SnapshotBondGroup         `json:"bond_groups"`
	CryptoGroups          []SnapshotCryptoGroup       `json:"crypto_groups"`
	Deposits              []SnapshotDeposit           `json:"deposits"`
	Insurances            []SnapshotInsurance         `json:"insurances"`
	CreditCards           []CreditCard                `json:"credit_cards"`
	Transactions          []SnapshotTransaction       `json:"stock_and_bond_transactions"`
	CryptoTransactions    []SnapshotCryptoTransaction `json:"crypto_transactions"`
	DepositTransactions   []DepositTransaction        `json:"deposit_transactions"`
	InsuranceTransactions []InsuranceTransaction      `json:"insurance_transactions"`
}
type SnapshotStockGroup struct {
	Name   string          `json:"name"`
	Weight int             `json:"weight"`
	Stocks []SnapshotStock `json:"stocks"`
}
type SnapshotStock struct {
	Name      string `json:"name"`
	Code      string `json:"code"`
	LotAmount int    `json:"lot_amount"`
	LotSize   int    `json:"lot_size"`
}
type SnapshotBondGroup struct {
	Name   string          `json:"name"`
	Weight int             `json:"weight"`
	Bonds  []SnapshotStock `json:"bonds"`
}
type SnapshotCryptoGroup struct {
	Name   string                `json:"name"`
	Weight int                   `json:"weight"`
	Assets []SnapshotCryptoAsset `json:"assets"`
}
type SnapshotCryptoAsset struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Platform       string  `json:"platform"`
	URL            string  `json:"url"`
	AnnualInterest float64 `json:"annual_interest"`
	Amount         string  `json:"amount"`
}
type SnapshotTransaction struct {
	ID             int64          `json:"id"`
	InstrumentCode string         `json:"instrument_code"`
	InstrumentType InstrumentType `json:"instrument_type"`
	LotAmount      int            `json:"lot_amount"`
	Buying         bool           `json:"buying"`
	LotSize        int            `json:"lot_size"`
	Price          float64        `json:"price"`
	Currency       string         `json:"currency"`
	Time           time.Time      `json:"time"`
}
type SnapshotCryptoTransaction struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	AssetName string    `json:"asset_name"`
	Platform  string    `json:"platform"`
	Amount    string    `json:"amount"`
	Buying    bool      `json:"buying"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}
type SnapshotDeposit struct {
	Name                     string  `json:"name"`
	BankName                 string  `json:"bank_name"`
	AmountRUB                float64 `json:"amount_rub"`
	AnnualInterest           float64 `json:"annual_interest"`
	StartDate                string  `json:"start_date"`
	EndDate                  string  `json:"end_date"`
	AcquisitionTransactionID int64   `json:"acquisition_transaction_id"`
}
type SnapshotInsurance struct {
	Name                     string  `json:"name"`
	CompanyName              string  `json:"company_name"`
	AmountRUB                float64 `json:"amount_rub"`
	AnnualInterest           float64 `json:"annual_interest"`
	StartDate                string  `json:"start_date"`
	EndDate                  string  `json:"end_date"`
	AcquisitionTransactionID int64   `json:"acquisition_transaction_id"`
}
