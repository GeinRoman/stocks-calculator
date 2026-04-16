package model

import "time"

type (
	Stock struct {
		Name      string
		Code      string
		GroupName string
		Price     float64
		Amount    int
	}

	Transaction struct {
		ProfileId int
		StockCode string
		Amount    int
		Buying    bool
		Price     float64
		Time      time.Time
	}

	FoundStock struct {
		Stock
		ErrMsg string
	}

	MoexFindResult struct {
		Stock
		Err error
	}

	MoexPriceResult struct {
		Price float64
		Err   error
	}
)
