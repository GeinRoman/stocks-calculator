package rebalance

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "rebalance",
		Short: "Calculate trades needed to rebalance your portfolio",
		Long: `Calculate which stocks to buy or sell to restore target allocation ratios.

Shows the specific trades (buy/sell actions with lot quantities) needed to
align your current portfolio with the target weights assigned to each group.

By default, rebalancing may involve both buying and selling. Use --nosell
to restrict rebalancing to purchases only (requires --deposit to add funds).

Deposit adds new capital to invest, while withdraw removes capital and shows
which positions to reduce.

In the end you will be asked whether you want to automatically update current stock info in stcalc`,
		Example: `  stcalc rebalance                           # rebalance with current holdings (no new money)
  stcalc rebalance --deposit 5000            # rebalance with 5,000 rubles of new capital
  stcalc rebalance --deposit 3000 --nosell   # rebalance with 3,000 rubles of new capital but prevents selling stocks
  stcalc rebalance --withdraw 20000          # Rebalance while withdrawing 20,000 rubles`,
		Args: cobra.ExactArgs(0),
		RunE: rebalanceCommand,
	}

	flags app.RebalanceOptions
)

func rebalanceCommand(cmd *cobra.Command, args []string) error {
	if flags.NoSell && flags.Deposit == 0 {
		return fmt.Errorf("--nosell flag requires --deposit flag. Unable to rebalance portfolio without either selling current stock or depositing new capital")
	}

	output, err := app.Rebalance(flags)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	nosell, deposit, withdraw := "nosell", "deposit", "withdraw"

	cmd.Flags().BoolVarP(&flags.NoSell, nosell, "n", false, "Only buy stocks, do not sell (requires --deposit)")
	cmd.Flags().IntVarP(&flags.Deposit, deposit, "d", 0, "Amount of rubles to add to portfolio")
	cmd.Flags().IntVarP(&flags.Withdraw, withdraw, "w", 0, "Amount of rubles to remove from portfolio")

	cmd.MarkFlagsMutuallyExclusive(deposit, withdraw)
	cmd.MarkFlagsMutuallyExclusive(nosell, withdraw)

	rootCmd.AddCommand(cmd)
}
