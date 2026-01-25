package commands

import (
	"log"
	"stocks_calculator/internal/cli/commands/add"
	deletecmd "stocks_calculator/internal/cli/commands/delete"
	"stocks_calculator/internal/cli/commands/login"
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
	login.Register(rootCmd)
	add.Register(rootCmd)
	deletecmd.Register(rootCmd)
	rebalance.Register(rootCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
