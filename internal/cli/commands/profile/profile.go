package profile

import (
	"fmt"
	"stocks_calculator/internal/cli/app"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "profile",
		Short: "Add, delete, or set a profile as default",
		Long: `Add, delete, or set a profile as the default.

Command without arguments will show avalible profiles and current default profile
If a profile with the given name does not exist, it will be created.
If it already exists, it will be set as the default profile.`,
		Example: `  stcalc profile                 # show avalible profiles and current default profile
  stcalc profile work            # create or set "work" as default
  stcalc profile work --remove   # remove "work" profile`,
		Args: cobra.MaximumNArgs(1),
		RunE: profileCommnad,
	}

	flags app.ProfileOptions
)

func profileCommnad(cmd *cobra.Command, args []string) error {
	var (
		output string
		err    error
	)

	if len(args) == 0 {
		if flags.Remove {
			return fmt.Errorf("To remove profile please specify [profile-name]")
		}

		flags.Info = true
		output, err = app.Profile("", &flags)
	} else {
		output, err = app.Profile(args[0], &flags)
	}

	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func Register(rootCmd *cobra.Command) {
	cmd.Flags().BoolVarP(&flags.Remove, "remove", "r", false, "Remove the profile with the specified name")

	rootCmd.AddCommand(cmd)
}
