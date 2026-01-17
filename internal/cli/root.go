package cli

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stcalc",
	Short: "stcalc (stocks calculator) is a tool to help manage stocks portfolio",
	Long:  "stcalc (stocks calculator) is a tool to help manage stocks portfolio ... [add something later]",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
