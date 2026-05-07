package stock

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "stock",
		Short: "Manage portfolio stocks and view stock information",
		Long: `Manage stocks in your portfolio.

When called without subcommands, displays information about all stocks
across all groups, including current prices, lot sizes, quantities held,
and total values.

Use subcommands to add or remove stocks from your portfolio.`,
		Example: `  stcalc stock                       # View all stocks in your portfolio
  stcalc stock add Yandex 3 PLZL 10  # Add stocks to portfolio
  stcalc stock remove Yandex 2       # Remove stocks from portfolio`,
		Args: cobra.ExactArgs(0),
		RunE: stock,
	}
)

func stock(cmd *cobra.Command, args []string) error {
	output, err := app.StockInfo()
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	cmd.AddCommand(addCmd)
	removeCmdFlags()
	cmd.AddCommand(removeCmd)

	rootCmd.AddCommand(cmd)
}
