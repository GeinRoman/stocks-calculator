package model

type (
	MoexFindResult struct {
		Stock
		Err error
	}

	MoexPriceResult struct {
		Price float64
		Err   error
	}
)
