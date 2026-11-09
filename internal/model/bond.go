package model

type Bond struct {
	Name            string   `json:"name"`
	Code            string   `json:"code"`
	GroupName       string   `json:"group_name"`
	CleanPrice      float64  `json:"clean_price"`      // RUB per bond, excluding accrued interest.
	AccruedInterest float64  `json:"accrued_interest"` // RUB per bond.
	MaturityDate    string   `json:"maturity_date"`    // YYYY-MM-DD.
	MaturityPayment *float64 `json:"maturity_payment"` // RUB principal per bond; nil for a non-single-payment schedule.
	LotAmount       int      `json:"lot_amount"`
	LotSize         int      `json:"lot_size"`
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
