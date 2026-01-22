package add

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "add",
		Short: "add a stock or a group",
		Long:  "add a stock or a group [add some later]",
		Args: cobra.MinimumNArgs(1),
		RunE:  addCommand,
	}

	flags app.AddOptions
)

func addCommand(cmd *cobra.Command, args []string) error {
	output, err := app.Add(&flags, args)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	stocFlagName := "stock"
	groupFlagName := "group"

	cmd.Flags().BoolVarP(&flags.Stock, stocFlagName, "s", false, "add stock or stocks to the portfolio")
	cmd.Flags().BoolVarP(&flags.Group, groupFlagName, "g", false, "add group or groups of stocks")

	cmd.MarkFlagsOneRequired(groupFlagName, stocFlagName)
	cmd.MarkFlagsMutuallyExclusive(groupFlagName, stocFlagName)

	rootCmd.AddCommand(cmd)
}
