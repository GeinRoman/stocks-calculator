package commands

import (
	"log"
	"stocks_calculator/internal/cli/app"
	"stocks_calculator/internal/cli/commands/connect"
	"stocks_calculator/internal/cli/commands/group"
	"stocks_calculator/internal/cli/commands/login"
	"stocks_calculator/internal/cli/commands/profile"
	"stocks_calculator/internal/cli/commands/rebalance"
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
	login.Register(rootCmd)
	profile.Register(rootCmd)
	group.Register(rootCmd)
	connect.Register(rootCmd)
	stock.Register(rootCmd)
	rebalance.Register(rootCmd)
}

func rootPreRunE(cmd *cobra.Command, args []string) error {
	cmds := strings.Split(cmd.CommandPath(), " ")
	// stcalc command without subcommands
	if len(cmds) < 2 {
		return nil
	}

	if cmds[1] == "connect" {
		return nil
	}

	if err := app.ValidateConnectionString(); err != nil {
		return err
	}

	if cmds[1] == "login" {
		return nil
	}

	if err := app.ValidateLogin(); err != nil {
		return err
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
