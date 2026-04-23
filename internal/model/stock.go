package model

import "time"

type (
	Stock struct {
		Name      string  `json:"name"`
		Code      string  `json:"code"`
		GroupName string  `json:"group_name"`
		Price     float64 `json:"price"`
		LotAmount int     `json:"lot_amount"`
		LotSize   int     `json:"lot_size"`
	}

	Transaction struct {
		ProfileId int       `json:"profile_id"`
		StockCode string    `json:"stock_code"`
		LotAmount int       `json:"lot_amount"`
		Buying    bool      `json:"buying"`
		LotSize   int       `json:"lot_size"`
		Price     float64   `json:"price"`
		Time      time.Time `json:"time"`
	}

	FoundStock struct {
		Stock  `json:"stock"`
		ErrMsg string `json:"err_msg"`
	}
)
