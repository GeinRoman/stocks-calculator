package stock

import "github.com/spf13/cobra"

var (
	cmd = &cobra.Command{
		Use:     "stock",
		Short:   "...",
		Long:    `...`,
		Example: `...`,
		Args:    cobra.ExactArgs(0),
	}
)

func Register(rootCmd *cobra.Command) {
	cmd.AddCommand(addCmd)
	removeCmdFlags()
	cmd.AddCommand(removeCmd)

	rootCmd.AddCommand(cmd)
}
