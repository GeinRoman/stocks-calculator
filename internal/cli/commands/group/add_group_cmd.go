package group

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:   "add <group-name>...",
		Short: "Add one or more groups to your portfolio",
		Long:  `Add groups (categories) to organize stocks in your portfolio.

Groups help you categorize stocks by sector, strategy, risk level, or any
other criteria that suits your investment approach. You can add multiple
groups at once by providing multiple names.`,
		Example: `  stcalc group add IT                 # add a new group
  stcalc group add Oil Retail         # add multiple groups`,
		Args: cobra.MinimumNArgs(1),
		RunE: add,
	}
)

func add(cmd *cobra.Command, args []string) error {
	output, err := app.Add(args)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
