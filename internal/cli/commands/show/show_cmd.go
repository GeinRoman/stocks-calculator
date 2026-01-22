package show

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "show",
		Short: "Print inforamtion about user's portfolio",
		Long:  "Print inforamtion about user's portfolio [add some later]",
		RunE:  showCommand,
	}

	flags app.ShowOptions
)

func showCommand(cmd *cobra.Command, args []string) error {
	output, err := app.Show(&flags)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	stocksFlagName := "stocks"
	groupFlagName := "groups"
	infoFlagName := "info"

	cmd.Flags().BoolVarP(&flags.Stocks, stocksFlagName, "s", false, "shows informaiton about owned stocks")
	cmd.Flags().BoolVarP(&flags.Groups, groupFlagName, "g", false, "shows informaiton about groups (stock categories)")
	cmd.Flags().StringVar(&flags.Info, infoFlagName, "", "shows informaiton about particular stock")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "show full information")

	cmd.MarkFlagsMutuallyExclusive(stocksFlagName, infoFlagName, groupFlagName)

	rootCmd.AddCommand(cmd)
}
