package profile

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "profile",
		Short: "log into remote server",
		Long:  "log into remote server [add long description]",
		Args:  cobra.ExactArgs(1),
		RunE:  profileCommnad,
	}

	flags app.ProfileOptions
)

func profileCommnad(cmd *cobra.Command, args []string) error {
	output, err := app.Profile(args[0], &flags)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	cmd.Flags().BoolVarP(&flags.NewAccount, "new", "n", false, "create new account with name specified")

	rootCmd.AddCommand(cmd)
}
