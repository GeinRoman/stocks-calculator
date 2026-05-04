package group

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	removeCmd = &cobra.Command{
		Use:   "remove <group-name>...",
		Short: "Remove one or more groups from your portfolio",
		Long: `Remove groups from your portfolio by name. 
You can remove multiple groups at once by providing multiple names.`,
		Example: `  stcalc group remove IT                 # remove group
  stcalc group remove Oil Retail         # remove multiple groups`,
		Args: cobra.MinimumNArgs(1),
		RunE: remove,
	}
)

func remove(cmd *cobra.Command, args []string) error {
	output, err := app.RemoveGroups(args)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
