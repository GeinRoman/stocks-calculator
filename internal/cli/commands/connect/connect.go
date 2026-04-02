package connect

import (
	"errors"
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{

		Use:   "connect",
		Short: "Set connection string",
		Long: `Configure the service connection settings.

Set the host address and optionally the port. When only the host is provided,
the default port 8989 is used. Run without arguments to view current settings.`,
		Example: `  stcalc connect                           # show current connection settings
  stcalc connect 192.168.0.1               # set host (uses default port 8989)
  stcalc connect api.example.com           # set host to domain name
  stcalc connect 8080 --set-port           # change port only (keeps current host)`,
		Args: cobra.MaximumNArgs(1),
		RunE: connectCommand,
	}

	flags app.ConnectOptions
)

func connectCommand(cmd *cobra.Command, args []string) error {
	var (
		output string
		err    error
	)

	if len(args) == 0 {
		if flags.Port {
			return errors.New("Please provide port number to set")
		}

		flags.Info = true
		output, err = app.Connect("", &flags)
	} else {
		output, err = app.Connect(args[0], &flags)
	}

	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	cmd.Flags().BoolVarP(&flags.Port, "set-port", "p", false, "Set port")

	rootCmd.AddCommand(cmd)
}
