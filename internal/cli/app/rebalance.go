package app

type RebalanceOptions struct {
	NoSell bool
}

func Rebalance(investment int, options *RebalanceOptions) (string, error) {
	return "rebalance", nil
}
