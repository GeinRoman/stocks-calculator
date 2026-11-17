package model

type (
	MoexFindResult struct {
		SearchResults []MoexSearchResult
		Err           error
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

	MoexBondFindResult struct {
		SearchResults []MoexBondSearchResult
		Err           error
	}

	MoexBondSearchResult struct {
		Bond
		Err error
	}

	MoexBondInfoResult struct {
		Bond
		Err error
	}
)
