package group

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	renameCmd = &cobra.Command{
		Use:   "rename <group-name-old | group-index> <group-name-new>",
		Short: "Rename group from your portfolio",
		Long: `Rename group from your portfolio. 
You can choose group to rename by it's name or index.`,
		Example: `  
  stcalc group rename Oil "Oil and Minerals"  # rename group
  stcalc group rename 1 "Oil and Minerals"    # rename group specified with index`,
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
