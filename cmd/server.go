package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/yoru9zine/gj"
)

var (
	serverLogDir string
)

func init() {
	ServerCmd.Flags().StringVarP(&serverLogDir, "logdir", "", "./log", "")
	RootCmd.AddCommand(ServerCmd)
}

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.MkdirAll(serverLogDir, 0755); err != nil {
			log.Fatalf("failed to create logdir `%s`: %s", serverLogDir, err)
		}
		srv := gj.NewAPIServer(serverLogDir)
		srv.Run(fmt.Sprintf(":%d", port))
	},
}
