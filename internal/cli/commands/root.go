package commands

import (
	"log"
	"stocks_calculator/internal/cli/commands/connect"
	deletecmd "stocks_calculator/internal/cli/commands/delete"
	"stocks_calculator/internal/cli/commands/group"
	"stocks_calculator/internal/cli/commands/profile"
	"stocks_calculator/internal/cli/commands/rebalance"
	"stocks_calculator/internal/cli/commands/show"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stcalc",
	Short: "stcalc (stocks calculator) is a tool to help manage stocks portfolio",
	Long:  "stcalc (stocks calculator) is a tool to help manage stocks portfolio ... [add something later]",
}

func init() {
	show.Register(rootCmd)
	profile.Register(rootCmd)
	deletecmd.Register(rootCmd)
	rebalance.Register(rootCmd)
	connect.Register(rootCmd)
	group.Register(rootCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
