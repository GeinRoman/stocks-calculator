package login

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{

		Use:   "login",
		Short: "Show current user, log in, or create a new user",
		Long:  `Display the currently logged-in user, log in as an existing user with --user, or create and log in as a new user by combining --user with --new.`,
		Example: `  stcalc login                       # Show current user
  stcalc login -u alice              # Log in as existing user
  stcalc login -u alice --new        # Create and log in as new user
  stcalc login -u alice -n           # Create and log in as new user`,
		Args: cobra.ExactArgs(0),
		RunE: loginCommand,
	}

	flags app.LoginOptions
)

func loginCommand(cmd *cobra.Command, args []string) error {
	if flags.New && flags.User == "" {
		return fmt.Errorf("To create new user specify user name with --user flag")
	}
	output, err := app.Login(flags)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	cmd.Flags().StringVarP(&flags.User, "user", "u", "", "Username to log in as")
	cmd.Flags().BoolVarP(&flags.New, "new", "n", false, "Create a new user with the name given by --user")

	rootCmd.AddCommand(cmd)
}
