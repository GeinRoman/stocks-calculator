package group

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	renameCmd = &cobra.Command{
		Use:   "rename <group-name-old> <group-name-new>",
		Short: "Rename group from your portfolio",
		Long:  `Rename group from your portfolio.`,
		Example: `  
  stcalc group rename Oil "Oil and Minerals"  # rename group`,
		Args: cobra.ExactArgs(2),
		RunE: rename,
	}
)

func rename(cmd *cobra.Command, args []string) error {
	output, err := app.RenameGroup(args[0], args[1])
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
