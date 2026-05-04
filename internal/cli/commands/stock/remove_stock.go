package stock

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	removeCmd = &cobra.Command{
		Use:   "remove <code> [lots] [code lots]...",
		Short: "Remove stocks from your portfolio",
		Long: `Remove one or more stocks from your portfolio with an optional number of lots.

You can specify stock by its code or code.

A lot represents the minimum tradable unit of a stock - the number of shares
that must be bought or sold together. Quantity is always specified in lots,
not individual shares. Defaults to 1 lot if not specified.

If removing fewer lots than currently held, the remaining lots stay in the
portfolio. If removing all lots (or more than held), the stock is completely
removed from the group. Aslo you can use "--all" flag to completely remove stock(s).`,
		Example: `  stcalc stock remove Yandex                           # Remove 1 lot (defaults to 1)
  stcalc stock remove Yandex 3                         # Remove 3 lots of a stock
  stcalc stock remove Yandex PLZL 10                   # Remove multiple stocks at once
  stcalc stock remove Yandex PLZL --all                # Remove all lots of stocks specified`,
		Args: cobra.MinimumNArgs(1),
		RunE: remove,
	}

	removeFlags app.RemoveStockOptions
)

func remove(cmd *cobra.Command, args []string) error {
	output, err := app.RemoveStock(removeFlags, args)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func removeCmdFlags() {
	removeCmd.Flags().BoolVar(&removeFlags.All, "all", false, "Remove all lots of stock(s) specified. (Has priority over number of lots)")
}
