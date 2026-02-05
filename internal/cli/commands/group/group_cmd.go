package group

import "github.com/spf13/cobra"

var (
	cmd = &cobra.Command{
		Use:   "group",
		Short: "...",
		Long:  `...`,
		Args:  cobra.MinimumNArgs(1),
		// RunE: groupCommand,
	}
)

// func groupCommand(cmd *cobra.Command, args []string) error {
//
// }

func Register(rootCmd *cobra.Command) {
	cmd.AddCommand(addCmd)
	cmd.AddCommand(removeCmd)

	rootCmd.AddCommand(cmd)
}
