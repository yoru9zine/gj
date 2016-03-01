package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/yoru9zine/gj"
)

func init() {
	RootCmd.AddCommand(RunCmd)
}

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run process",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("process file required")
		}
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("failed to open file %s: %s", args[0], err)
		}
		defer f.Close()
		client := gj.NewClient(fmt.Sprintf("http://localhost:%d", port))
		pid, err := client.Create(f)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		if err := client.Start(pid); err != nil {
			log.Fatalf("failed to start %s: %s", pid, err)
		}
		fmt.Printf("%s\n", pid)
	},
}
