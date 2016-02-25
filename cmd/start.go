package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/yoru9zine/gj"
)

func init() {
	RootCmd.AddCommand(StartCmd)
}

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start process",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("pid required")
		}
		client := gj.NewClient(fmt.Sprintf("http://localhost:%d", port))
		if err := client.Start(args[0]); err != nil {
			log.Fatalf("error: %s", err)
		}
		fmt.Println("started")
	},
}
