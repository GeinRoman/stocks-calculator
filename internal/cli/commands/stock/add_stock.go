package stock

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:   "add <name> [lots] [name lots]...",
		Short: "Add stocks to your portfolio",
		Long: `Add one or more stocks to your portfolio with an optional number of lots.

You can specify stock by its name or code.

A lot represents the minimum tradable unit of a stock - the number of shares
that must be bought or sold together. Quantity is always specified in lots,
not individual shares. Defaults to 1 lot if not specified.

If a stock already exists in the portfolio, the specified lots will be
appended to the current holding.

You will be prompted to select group interactively for new stocks.`,
		Example: `  stcalc stock add Yandex                           # Add a single stock (defaults to 1 lot)
  stcalc stock add Yandex 3                         # Add 3 lots of a stock
  stcalc stock add Yandex PLZL 10                   # Add multiple stocks at once`,
		Args: cobra.MinimumNArgs(1),
		RunE: add,
	}

	addFlags app.AddStockOptions
)

func add(cmd *cobra.Command, args []string) error {
	output, err := app.AddStock(addFlags, args)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
