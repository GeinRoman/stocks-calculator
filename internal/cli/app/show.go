package app

type ShowOptions struct {
	Verbose bool
	Stocks  bool
	Groups bool
	Info    string
}

func Show(flags *ShowOptions) (string, error) {
	switch {
	case flags.Stocks:
		return showStocks(flags.Verbose)
	case flags.Groups:
		return showGroups(flags.Verbose)
	case flags.Info != "":
		return showStockInfo(flags.Info, flags.Verbose)
	default:
		return showPortfolio(flags.Verbose)
	}
}

func showPortfolio(verbose bool) (string, error) {
	//http request here ...
	var output string
	if verbose {
		output = "showPortfolioVerbose"
	} else {
		output = "showPortfolioShort"
	}
	return output, nil
}

func showStocks(verbose bool) (string, error) {
	//http request here ...
	var output string
	if verbose {
		output = "showStocksVerbose"
	} else {
		output = "showStocksShort"
	}
	return output, nil
}

func showStockInfo(stock string, verbose bool) (string, error) {
	//http request here ...
	var output string
	if verbose {
		output = "showStockInfoVerbose " + stock
	} else {
		output = "showStockInfoShort " + stock
	}
	return output, nil
}

func showGroups(verbose bool) (string, error) {
	//http request here ...
	var output string
	if verbose {
		output = "showGroupsVerbose"
	} else {
		output = "showGroupsShort"
	}
	return output, nil
}
