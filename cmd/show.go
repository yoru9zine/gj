package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/yoru9zine/gj"
)

func init() {
	RootCmd.AddCommand(ShowCmd)
}

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show process detail",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("pid required")
		}
		client := gj.NewClient(fmt.Sprintf("http://localhost:%d", port))
		model, err := client.Show(args[0])
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		b, err := json.MarshalIndent(model, "", "  ")
		if err != nil {
			log.Fatalf("invalid response: %s\n%s", err, b)
		}
		fmt.Printf("%s\n", b)
	},
}
