package commands

import (
	"log"
	"stocks_calculator/internal/cli/app"
	"stocks_calculator/internal/cli/commands/connect"
	"stocks_calculator/internal/cli/commands/group"
	"stocks_calculator/internal/cli/commands/profile"
	"stocks_calculator/internal/cli/commands/stock"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "stcalc",
	Short:             "stcalc (stocks calculator) is a tool to help manage stocks portfolio",
	Long:              "stcalc (stocks calculator) is a tool to help manage stocks portfolio ... [add something later]",
	PersistentPreRunE: rootPreRunE,
}

func init() {
	profile.Register(rootCmd)
	group.Register(rootCmd)
	connect.Register(rootCmd)
	stock.Register(rootCmd)

	// rebalance.Register(rootCmd)
}

func rootPreRunE(cmd *cobra.Command, args []string) error {
	cmds := strings.Split(cmd.CommandPath(), " ")
	if len(cmds) >= 2 && cmds[1] != "profile" && cmds[1] != "connect" {
		return app.ValidateUserConfig()
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
