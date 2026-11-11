package model

type CryptoGroup struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

type RenameCryptoGroupBody struct {
	NameOld string `json:"name_old"`
	NameNew string `json:"name_new"`
}

type CryptoAsset struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	GroupName      string  `json:"group_name"`
	Platform       string  `json:"platform"`
	URL            string  `json:"url"`
	AnnualInterest float64 `json:"annual_interest"`
	Amount         float64 `json:"amount"`
}

// CryptoAssetUpdate replaces metadata without changing the holding's amount.
type CryptoAssetUpdate struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Platform       string  `json:"platform"`
	URL            string  `json:"url"`
	AnnualInterest float64 `json:"annual_interest"`
}

type CryptoOperation struct {
	CryptoAsset
	Price Money `json:"price"`
}
