package rebalance

import (
	"errors"
	"fmt"
	"stocks_calculator/internal/cli/app"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "rebalance",
		Short: "Print information about which stocks to buy or sell to properly ballance portfolio",
		Long:  "Print information about which stocks to buy or sell to properly ballance portfolio [add some later]. Specify an amount of rubles to invest as agrument (without it will be assumed that investment is 0)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  rebalanceCommand,
	}

	flags app.RebalanceOptions
)

func rebalanceCommand(cmd *cobra.Command, args []string) error {
	investment := 0
	if len(args) == 1 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return errors.New("Cannot convert \"" + args[0] + "\" to an integer")
		}
		investment = num
	}

	if flags.NoSell && len(args) == 0 {
		return errors.New("Cannot perform rebalance without selling existing stocks and without new investments")
	}

	output, err := app.Rebalance(investment, &flags)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	cmd.Flags().BoolVarP(&flags.NoSell, "nosell", "n", false, "Rebalance portfolio in best way possible without selling any of currently aquired stocks")

	rootCmd.AddCommand(cmd)
}
