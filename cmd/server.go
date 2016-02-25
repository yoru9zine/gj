package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoru9zine/gj"
)

func init() {
	RootCmd.AddCommand(ServerCmd)
}

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Run: func(cmd *cobra.Command, args []string) {
		srv := gj.NewAPIServer()
		srv.Run(fmt.Sprintf(":%d", port))
	},
}
