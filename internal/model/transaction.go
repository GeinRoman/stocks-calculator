package model

import "time"

type InstrumentType string

const (
	InstrumentStock InstrumentType = "stock"
	InstrumentBond  InstrumentType = "bond"
)

type Transaction struct {
	ProfileId      int            `json:"profile_id"`
	InstrumentCode string         `json:"instrument_code"`
	InstrumentType InstrumentType `json:"instrument_type"`
	LotAmount      int            `json:"lot_amount"`
	Buying         bool           `json:"buying"`
	LotSize        int            `json:"lot_size"`
	Price          float64        `json:"price"`
	Currency       string         `json:"currency"`
	Time           time.Time      `json:"time"`
}
