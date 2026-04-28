package model

type (
	MoexFindResult struct {
		SearchResults []Stock
		Err           error
	}

	MoexInstrumentInfoResult struct {
		Price   float64
		LotSize int
		Err     error
	}
)
