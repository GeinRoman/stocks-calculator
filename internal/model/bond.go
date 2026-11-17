package model

type Bond struct {
	Name            string `json:"name"`
	Code            string `json:"code"`
	ISIN            string `json:"isin"`
	GroupName       string `json:"group_name"`
	CleanPrice      Money  `json:"clean_price"`      // Per bond, excluding accrued interest.
	AccruedInterest Money  `json:"accrued_interest"` // Per bond, in settlement currency.
	MaturityDate    string `json:"maturity_date"`    // YYYY-MM-DD.
	MaturityPayment *Money `json:"maturity_payment"` // Principal per bond; nil for multiple payments.
	LotAmount       int    `json:"lot_amount"`
	LotSize         int    `json:"lot_size"`
}

type BondOperation struct {
	Name      string `json:"name"`
	Code      string `json:"code"`
	GroupName string `json:"group_name"`
	LotAmount int    `json:"lot_amount"`
	LotSize   int    `json:"lot_size"`
	Price     Money  `json:"price"` // Paid or received per bond, including accrued interest.
}

type BondGroup struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

type RenameBondGroupBody struct {
	NameOld string `json:"name_old"`
	NameNew string `json:"name_new"`
}

type FoundBond struct {
	SearchResults []Bond `json:"search_results"`
	ErrMsg        string `json:"err_msg"`
}
