package model

type PersonalAssets struct {
	Deposits    []Deposit    `json:"deposits"`
	CreditCards []CreditCard `json:"credit_cards"`
	Insurances  []Insurance  `json:"insurances"`
}

type AllInfo struct {
	Bonds          []Bond         `json:"bonds"`
	PersonalAssets PersonalAssets `json:"personal_assets"`
	CryptoAssets   []CryptoAsset  `json:"crypto_assets"`
	Stocks         []Stock        `json:"stocks"`
	Profiles       []Profile      `json:"profiles"`
}
