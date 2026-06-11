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
	Use:   "stcalc",
	Short: "stcalc (stocks calculator) is a tool to help manage stocks portfolio",
	Long: `stcalc is a CLI tool for managing and rebalancing stock portfolios.

The application allows users to organize stocks into groups, assign target
weights to groups, and calculate how to rebalance a portfolio while preserving
desired allocation ratios.

stcalc works with a backend service that stores portfolio data, retrieves
stock prices, and synchronizes user profiles across multiple devices.

To start:
1. add connection string with       #  stcalc connect
2. login (or create new user)       #  stcalc login
3. create profile                   #  stcalc profile
4. add and weight groups            #  stcalc group
5. add stocks to groups             #  stcalc stock
6. optimize portfolio               #  stcalc rebalance`,
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
