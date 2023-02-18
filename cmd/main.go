package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "cpngo",
	Short: "cpngo is command line interface for working with Colored Petri Nets",
	Long:  "cpngo is command line interface for working with Colored Petri Nets",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("hello world!")
		return nil
	},
}
