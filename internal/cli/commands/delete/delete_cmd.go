package deletecmd

import (
	"errors"
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "delete",
		Short: "delete a stock or a group",
		Long:  "delete a stock or a group [add some later]",
		RunE:  deleteCommand,
	}

	flags app.DeleteOptions
)

func deleteCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && !flags.All {
		return errors.New("Please specify what to delete")
	}

	if len(args) != 0 && flags.All {
		return errors.New("All flag is specified alongside delete arguments")
	}

	output, err := app.Delete(&flags, args)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	stocFlagName := "stock"
	groupFlagName := "group"

	cmd.Flags().BoolVarP(&flags.Stock, stocFlagName, "s", false, "delete stock or stocks to the portfolio")
	cmd.Flags().BoolVarP(&flags.Group, groupFlagName, "g", false, "delete group or groups of stocks")
	cmd.Flags().BoolVar(&flags.All, "all", false, "delete all groups or stocks")

	cmd.MarkFlagsOneRequired(groupFlagName, stocFlagName)
	cmd.MarkFlagsMutuallyExclusive(groupFlagName, stocFlagName)

	rootCmd.AddCommand(cmd)
}
