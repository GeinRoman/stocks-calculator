package login

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "login",
	Short: "log into remote server",
	Long:  "log into remote server [add long description]",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Please, enter password for %s:\n", args[0])

		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if password[:len(password)-1] == "123" {
			fmt.Println("Logged in successfuly!")
		} else {
			fmt.Println(password)
			fmt.Println("No user with such credatials found!")
		}

		return nil
	},
}
