package group

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "group",
		Short: "Manage portfolio groups and view group information",
		Long: `Manage groups (categories) in your portfolio.

Groups help organize stocks by sector, strategy, risk level, or any other
criteria. When called without subcommands, displays information about all
existing groups including their weights and stock assignments.

Use subcommands to add, remove, rename groups, or set their weights.`,
		Example: `  stcalc group       # Display groups' information`,
		Args:    cobra.ExactArgs(0),
		RunE:    groupInfo,
	}
)

func groupInfo(cmd *cobra.Command, args []string) error {
	output, err := app.SprintGroupInfo()
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	cmd.AddCommand(addCmd)
	cmd.AddCommand(removeCmd)
	cmd.AddCommand(renameCmd)
	cmd.AddCommand(weightCmd)

	rootCmd.AddCommand(cmd)
}
