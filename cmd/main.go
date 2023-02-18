package main

import (
	"fmt"
	"os"

	"github.com/adamlouis/cpngo/cpngo"
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
		n, err := cpngo.NewNet(
			[]cpngo.Place{
				{ID: "p1"},
				{ID: "p2"},
				{ID: "p3"},
				{ID: "p4"},
				{ID: "p5"},
			},
			[]cpngo.Transition{
				{ID: "t1"},
				{ID: "t2"},
				{ID: "t3"},
				{ID: "t4"},
			},
			[]cpngo.InputArc{
				{ID: "p1t1", FromID: "p1", ToID: "t1"},
				{ID: "p2t2", FromID: "p2", ToID: "t2"},
				{ID: "p3t3", FromID: "p3", ToID: "t3"},
				{ID: "p4t4", FromID: "p4", ToID: "t4"},
			},
			[]cpngo.OutputArc{
				{ID: "t1p2", FromID: "t1", ToID: "p2"},
				{ID: "t1p3", FromID: "t1", ToID: "p3"},
				{ID: "t2p4", FromID: "t2", ToID: "p4"},
				{ID: "t3p4", FromID: "t3", ToID: "p4"},
				{ID: "t4p5", FromID: "t4", ToID: "p5"},
			},
			[]cpngo.Token{
				{ID: "t1", PlaceID: "p1", Color: "foobar"},
			},
		)
		if err != nil {
			return err
		}

		srv := &server.Server{
			Net: n,
		}
		return srv.Serve()
	},
}
