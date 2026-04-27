package model

type (
	MoexFindResult struct {
		Stock
		Err error
	}

	MoexInstrumentInfoResult struct {
		Price   float64
		LotSize int
		Err     error
	}
)
