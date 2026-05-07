package model

type (
	MoexFindResult struct {
		SearchResults []MoexSearchResult
		Err error
	}

	MoexSearchResult struct {
		Stock
		Err error
	}

	MoexInstrumentInfoResult struct {
		Price   float64
		LotSize int
		Err     error
	}
)
