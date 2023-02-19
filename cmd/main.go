package main

import (
	"fmt"
	"os"

	"github.com/adamlouis/cpngo/internal/server"
	"github.com/spf13/cobra"
)

func main() {
	Execute()
}

func init() {
	rootCmd.AddCommand(serveCommand)
}

func Execute() {

	serveCommand.Flags().IntP("port", "p", 8888, "port to run the server on")

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

var serveCommand = &cobra.Command{
	Short: "run a CPN webserver",
	Use:   "serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}
		return (&server.Server{
			Port: port,
		}).Serve()
	},
}
