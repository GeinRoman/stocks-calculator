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

A lot represents the minimum tradable unit of a stock - the number of shares
that must be bought or sold together. Quantity is always specified in lots,
not individual shares. Defaults to 1 lot if not specified.

If a stock already exists in the portfolio, the specified lots will be
appended to the current holding.

A group can be assigned to all added stocks using the --group flag,
accepting either a group name or its index. If no group is provided,
you will be prompted to select one interactively.`,
		Example: `  
  stcalc stock add Yandex                           # Add a single stock (defaults to 1 lot)
  stcalc stock add Yandex 3                         # Add 3 lots of a stock
  stcalc stock add Yandex 3 --group "Good Stocks"   # Add stock to a specific group by index
  stcalc stock add Yandex PLZL 10                   # Add multiple stocks at once
  stcalc stock add Yandex 3 PLZL 10 --group 1       # Add multiple stocks and assign them to a group`,
		Args: cobra.MinimumNArgs(1),
		RunE: add,
	}

	flags app.AddStockOptions
)

func add(cmd *cobra.Command, args []string) error {
	output, err := app.AddStock(flags, args)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func addCmdFlags() {
	addCmd.Flags().StringVarP(&flags.Group, "group", "g", "", "Specify group name or index")
}
