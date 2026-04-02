package group

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	weightCmd = &cobra.Command{
		Use:     "weight",
		Short:   "Weight group from your portfolio.",
		Long:    "Weight group from your portfolio.",
		Example: `  stcalc group weight              # weight groups`,
		Args:    cobra.ExactArgs(0),
		RunE:    weight,
	}
)

func weight(cmd *cobra.Command, args []string) error {
	output, err := app.Weight()
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
