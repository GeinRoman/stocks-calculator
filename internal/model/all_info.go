package model

type PersonalAssets struct {
	Deposits    []Deposit    `json:"deposits"`
	CreditCards []CreditCard `json:"credit_cards"`
	Insurances  []Insurance  `json:"insurances"`
}

type AllInfo struct {
	Bonds          []Bond         `json:"bonds"`
	BondGroups     []BondGroup    `json:"bond_groups"`
	PersonalAssets PersonalAssets `json:"personal_assets"`
	CryptoAssets   []CryptoAsset  `json:"crypto_assets"`
	CryptoGroups   []CryptoGroup  `json:"crypto_groups"`
	Stocks         []Stock        `json:"stocks"`
	StockGroups    []StockGroup   `json:"stock_groups"`
	Profiles       []Profile      `json:"profiles"`
}
