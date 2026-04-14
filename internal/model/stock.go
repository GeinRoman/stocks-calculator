package model

type (
	Stock struct {
		Name      string
		Code      string
		GroupName string
		Price     float64
		LotAmount int
		LotSize   int
	}
)
