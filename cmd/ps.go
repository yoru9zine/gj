package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/yoru9zine/gj"
)

func init() {
	RootCmd.AddCommand(PSCmd)
}

var PSCmd = &cobra.Command{
	Use:   "ps",
	Short: "Show process list",
	Run: func(cmd *cobra.Command, args []string) {
		client := gj.NewClient(fmt.Sprintf("http://localhost:%d", port))
		models, err := client.PS()
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		for _, proc := range models {
			fmt.Printf("%s\t%s\n", proc.ID, proc.Name)
		}
	},
}
