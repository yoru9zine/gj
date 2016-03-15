package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/yoru9zine/gj"
)

func init() {
	RootCmd.AddCommand(LogsCmd)
}

var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show process log",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("pid required")
		}
		client := gj.NewClient(fmt.Sprintf("http://localhost:%d", port))
		logstring, err := client.Log(args[0])
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		fmt.Printf(logstring)
	},
}
