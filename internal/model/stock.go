package model

import "time"

type (
	Stock struct {
		Name      string
		Code      string
		GroupName string
		Price     float64
		LotAmount int
		LotSize   int
	}

	Transaction struct {
		ProfileId int
		StockCode string
		LotAmount int
		Buying    bool
		LotSize   int
		Price     float64
		Time      time.Time
	}

	FoundStock struct {
		Stock
		ErrMsg string
	}

	MoexResult struct {
		Stock
		Err error
	}
)
